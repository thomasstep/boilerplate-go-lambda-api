package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	// "github.com/lestrrat-go/httprc"
	// "github.com/lestrrat-go/jwx/v2/jwk"
	// "github.com/lestrrat-go/jwx/v2/jwt"
	"go.uber.org/zap"
)

// Helper function to generate an IAM policy
func generatePolicy(principalId string, effect string, resource string) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}

	// Optional output with custom properties of the String, Number or Boolean type.
	// Add userId string to function parameters
	// authResponse.Context = map[string]interface{}{
	// 	"userId": userId,
	// }
	return authResponse
}

func handleRequest(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	// Get resource before determining if user is allowed or not
	// first two pieces are apiGatewayArn and stage
	methodArnPieces := strings.Split(event.MethodArn, "/")
	// without doing this, the user only gains access to the single resource and
	// method. access to the rest of the api will be denied
	var apiStageArn string
	if len(methodArnPieces) >= 2 {
		apiStageArn = fmt.Sprintf("%s/%s/*", methodArnPieces[0], methodArnPieces[1])
	} else {
		logger.Error(
			"Could not parse method ARN",
			zap.String("methodArn", event.MethodArn),
		)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Error: Could not parse method ARN")
	}

	capHeader := event.Headers["Authorization"]
	lowerHeader := event.Headers["authorization"]
	header := capHeader
	if capHeader == "" {
		header = lowerHeader
	}
	headerPieces := strings.Split(header, " ")
	var token string
	if len(headerPieces) == 2 {
		token = headerPieces[1]
	}

	if token == "" {
		logger.Error(
			"Could not get token from headers",
			zap.Any("headers", event.Headers),
		)
		return generatePolicy("user", "Deny", apiStageArn), nil
	}

	return generatePolicy("user", "Allow", apiStageArn), nil

	// Get JWKS
	// jwksUrl := config.JwksUrl
	// if jwksUrl == "" {
	// 	return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Error: JWKS URL is not available")
	// }

	// whitelist := httprc.NewMapWhitelist()
	// whitelist.Add(jwksUrl)
	// // TODO would ideally like to cache this and only try to connect as a fallback
	// keySet, keySetErr := jwk.Fetch(context.TODO(), jwksUrl, jwk.WithFetchWhitelist(whitelist))
	// if keySetErr != nil {
	// 	logger.Error(
	// 		"Could not get JWKS",
	// 		zap.String("URL", jwksUrl),
	// 	)
	// 	return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Error: Could not get JWKS")
	// }

	// // Get JWK
	// jwkId := config.JwkId
	// if jwksUrl == "" {
	// 	return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Error: JWK ID is not available")
	// }

	// key, exists := keySet.LookupKeyID(jwkId)
	// if exists != true {
	// 	logger.Error(
	// 		"Could not find KID",
	// 		zap.Any("jwks", keySet),
	// 		zap.String("jwkId", jwkId),
	// 	)
	// 	return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Error: Could not find KID")
	// }

	// // Verify JWT with JWK
	// verifiedToken, parseErr := jwt.Parse([]byte(token), jwt.WithKey(key.Algorithm(), key))
	// if parseErr != nil {
	// 	logger.Error(
	// 		"Failed to verify JWT",
	// 		zap.Error(parseErr),
	// 	)

	// 	return generatePolicy("user", "Deny", apiStageArn, ""), nil
	// }

	// return generatePolicy("user", "Allow", apiStageArn, verifiedToken.Subject()), nil
}

func main() {
	lambda.Start(handleRequest)
}
