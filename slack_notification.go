package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var (
	httpClient *http.Client
)

func init() {
	httpClient = http.DefaultClient
}

type notification struct {
	Text string `json:"text"`
}

// SlackAlert ...
func SlackAlert(text string) {
	if slackWebHookUrl == "" {
		return
	}

	// send asynchronously
	go func() {
		notf := &notification{
			Text: "[cloud_run] " + text,
		}
		b, _ := json.Marshal(notf)

		data := url.Values{}
		data.Set("payload", string(b))

		req, _ := http.NewRequest(http.MethodPost, slackWebHookUrl, strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		if _, err := httpClient.Do(req); err != nil {
			log.Printf("error in sending slack notification, err %v\n", err)
		}
	}()
}
