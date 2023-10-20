package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/thomasstep/giphy-livechat-api/internal/adapters"
)

func handleRequest(ctx context.Context, snsEvent events.SNSEvent) {
	for _, record := range snsEvent.Records {
		snsRecord := record.SNS
		snsMessage := snsRecord.Message
		var message adapters.EventActionEvent
		unmarshalErr := json.Unmarshal([]byte(snsMessage), &message)
		if unmarshalErr != nil {
			logger.Error(unmarshalErr.Error())
		}

		logic(message.Entity)
	}
}

func main() {
	lambda.Start(handleRequest)
}
