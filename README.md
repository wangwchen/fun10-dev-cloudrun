## Deployment

```bash
# example
gcloud run deploy <your-service-name> \
 --allow-unauthenticated --source . \
 --region europe-west3 --cpu 1 --memory 128Mi \
 --timeout=3s --concurrency 80  --min-instances 1 \
 --max-instances 1000 \
 --update-env-vars GAME_ID=<your-game>,ENVIRONMENT=<your-env>,SIGNATURE_KEY=<your-key>,PROJECT_ID=<your-project-id>,TOPIC_NAME=<your-pubsub-topic>,SLACK_WEB_HOOK_URL=<your-slack-url>

# funcom Dev
gcloud run deploy funcom-cloudrun-seabass-dev-europe-west3 \
 --allow-unauthenticated --source . \
 --region europe-west3 --cpu 1 --memory 128Mi \
 --timeout=3s --concurrency 80  --min-instances 1 \
 --max-instances 1000 \
 --update-env-vars GAME_ID=seabass,ENVIRONMENT=dev,SIGNATURE_KEY=a1b2c3d4e5f6g7h8,PROJECT_ID=data-platform-2,TOPIC_NAME=funcom-seabass-dev,SLACK_WEB_HOOK_URL=https://hooks.slack.com/services/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

# funcom StressTest
gcloud run deploy funcom-cloudrun-seabass-stresstest-europe-west3 \
 --allow-unauthenticated --source . \
 --region europe-west3 --cpu 1 --memory 128Mi \
 --timeout=3s --concurrency 80  --min-instances 1 \
 --max-instances 1000 \
 --update-env-vars GAME_ID=seabass,ENVIRONMENT=stresstest,SIGNATURE_KEY=a1b2c3d4e5f6g7h8,PROJECT_ID=data-platform-2,TOPIC_NAME=funcom-seabass-stresstest,SLACK_WEB_HOOK_URL=https://hooks.slack.com/services/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

# funcom march playtest
gcloud run deploy funcom-cloudrun-seabass-marchpt-europe-west3 \
 --allow-unauthenticated --source . \
 --region europe-west3 --cpu 1 --memory 128Mi \
 --timeout=3s --concurrency 80  --min-instances 1 \
 --max-instances 1000 \
 --update-env-vars GAME_ID=seabass,ENVIRONMENT=marchpt,SIGNATURE_KEY=a1b2c3d4e5f6g7h8,PROJECT_ID=data-platform-2,TOPIC_NAME=funcom-seabass-marchpt,SLACK_WEB_HOOK_URL=https://hooks.slack.com/services/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
```