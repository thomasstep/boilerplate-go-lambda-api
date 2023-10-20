package main

import (
	"context"
	"encoding/json"

	"github.com/thomasstep/giphy-livechat-api/internal/common"
	"github.com/thomasstep/giphy-livechat-api/internal/types"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type ResponseStructure struct {
	Id string `json:"id"`
}

func lambdaAdapter(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	entityId := request.PathParameters["entityId"]
	var body types.EntityUpdates
	unmarshalErr := json.Unmarshal([]byte(request.Body), &body)
	if unmarshalErr != nil {
		panic(unmarshalErr)
	}

	entity, err := logic(entityId, body)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	jsonBody, marshalErr := json.Marshal(entity)
	if marshalErr != nil {
		return events.APIGatewayProxyResponse{}, marshalErr
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonBody),
	}, err
}

func getLambdaHandler() types.HandlerSignature {
	wrappedLambdaAdapter := common.LamdbaWrapper(lambdaAdapter)
	return wrappedLambdaAdapter
}

func main() {
	lambda.Start(getLambdaHandler())
}
