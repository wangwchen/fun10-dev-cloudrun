package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	//"strconv"

	"cloud.google.com/go/pubsub"
)

const (
	port = "8080"

	paramSignature = "signature"

	envGameID          = "GAME_ID"
	envEnvironment     = "ENVIRONMENT"
	envSignatureKey    = "SIGNATURE_KEY"
	envProjectID       = "PROJECT_ID"
	envTopicName       = "TOPIC_NAME"
	envErrorTopicName  = "ERROR_TOPIC_NAME"
	envSlackWebHookUrl = "SLACK_WEB_HOOK_URL"

	attributeTableName    = "ATTR_TABLE_NAME"
	attributeErrorMessage = "ATTR_ERROR_MESSAGE"
)

var (
	ctx = context.Background()

	signatureKey string

	slackWebHookUrl string

	pubsubTopic    *pubsub.Topic
	errPubsubTopic *pubsub.Topic
)

type requestStruct struct {
	LogName string          `json:"log_name"`
	LogData json.RawMessage `json:"log_data"`
}

type responseStruct struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data,omitempty"`
}

func main() {
	// get game id from env, as part of the uri, <gameid>/<env>/report
	gameID := os.Getenv(envGameID)
	if gameID == "" {
		log.Fatalf("env %s is required", envGameID)
	}
	// get environment from env, as part of the uri, <gameid>/<env>/report
	env := os.Getenv(envEnvironment)
	if env == "" {
		log.Fatalf("env %s is required", envEnvironment)
	}
	//if _, err := strconv.Atoi(gameID); err != nil {
	//		log.Fatalf("env %s is invalid, err %v", envGameID, err)
	//}
	log.Printf("service for game id %s,env %s\n", gameID, env)

	// get signature key from env, for signature validation
	signatureKey = os.Getenv(envSignatureKey)

	// get slack web hook url, for alert
	slackWebHookUrl = os.Getenv(envSlackWebHookUrl)

	// get project id from env
	projectID := os.Getenv(envProjectID)
	if projectID == "" {
		log.Fatalf("env %s is required", envProjectID)
	}
	log.Printf("service for project id %s\n", projectID)

	// init pubsubClient
	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("err in new pubsub client, err %v", err)
	}

	// get pubsub topic name from env
	topicName := os.Getenv(envTopicName)
	if topicName == "" {
		log.Fatalf("env %s is required", envTopicName)
	}
	log.Printf("service for pubsub topic %s\n", topicName)
	pubsubTopic = pubsubClient.Topic(topicName)

	// get error pubsub topic name from env
	errorTopicName := os.Getenv(envErrorTopicName)
	if errorTopicName != "" {
		log.Printf("service for error pubsub topic %s\n", errorTopicName)
		errPubsubTopic = pubsubClient.Topic(errorTopicName)
	}

	log.Println("starting server...")
	http.HandleFunc(fmt.Sprintf("/%s/%s/report", gameID, env), handler)

	log.Printf("listening on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// read body from request
	rsBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeResponse(w, &responseStruct{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		})
		return
	}

	// validate signature if enabled
	if signatureKey != "" {
		var signature string
		// get signature from header
		signatures, ok := r.Header[paramSignature]
		if ok && len(signatures) > 0 {
			signature = signatures[0]
		}

		// get signature from uri
		if signature == "" {
			signature = r.URL.Query().Get(paramSignature)
		}

		if signature == "" {
			writeResponse(w, &responseStruct{
				Code: http.StatusUnauthorized,
				Msg:  "signature is required",
			})
			return
		}

		// compared with signature calculated
		if signature != fmt.Sprintf("%x", md5.Sum([]byte(string(rsBody)+signatureKey))) {
			writeResponse(w, &responseStruct{
				Code: http.StatusForbidden,
				Msg:  "signature is invalid",
				Data: string(rsBody),
			})
			return
		}
	}

	var reqBody requestStruct
	if err := json.Unmarshal(rsBody, &reqBody); err != nil {
		writeResponse(w, &responseStruct{
			Code: http.StatusBadRequest,
			Msg:  err.Error(),
		})

		// push to error topic
		pushToErrorTopic(rsBody, "invalid msg: invalid json")

		// send slack alert
		SlackAlert(fmt.Sprintf("received invalid msg: %s", string(rsBody)))
		return
	}

	// some validations
	if reqBody.LogName == "" {
		writeResponse(w, &responseStruct{
			Code: http.StatusBadRequest,
			Msg:  "log_name is required",
		})

		// push to error topic
		pushToErrorTopic(rsBody, "invalid msg: log_name is required")

		// send slack alert
		SlackAlert(fmt.Sprintf("received invalid msg: %s, log_name is missing", string(rsBody)))
		return
	}
	if len(reqBody.LogData) < 2 {
		writeResponse(w, &responseStruct{
			Code: http.StatusBadRequest,
			Msg:  "log_data is required",
		})

		// push to error topic
		pushToErrorTopic(rsBody, "invalid msg: log_data is required")

		// send slack alert
		SlackAlert(fmt.Sprintf("received invalid msg: %s, log_data is missing", string(rsBody)))
		return
	}

	// forwards to pubsub topic
	result := pubsubTopic.Publish(ctx, &pubsub.Message{
		Data: reqBody.LogData,
		Attributes: map[string]string{
			attributeTableName: reqBody.LogName,
		},
	})

	// blocking
	if _, err := result.Get(ctx); err != nil {
		writeResponse(w, &responseStruct{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		})

		// send slack alert
		SlackAlert(fmt.Sprintf("failed to publish msg to pubsub, err %v", err))
		return
	}

	writeResponse(w, &responseStruct{
		Code: http.StatusOK,
		Msg:  "ok",
	})
}

func pushToErrorTopic(data []byte, errMsg string) {
	if errPubsubTopic == nil {
		log.Println("error topic is not setup")
		return
	}

	result := pubsubTopic.Publish(ctx, &pubsub.Message{
		Data: data,
		Attributes: map[string]string{
			attributeErrorMessage: errMsg,
		},
	})

	if _, err := result.Get(ctx); err != nil {
		log.Printf("error in publishing message to error topic, err %v\n", err)
	}
}

func writeResponse(w http.ResponseWriter, res *responseStruct) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
