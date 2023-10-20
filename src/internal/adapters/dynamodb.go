package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"

	"github.com/thomasstep/giphy-livechat-api/internal/types"
)

/*
 * The main functions to be used from this package are intended to be
 * the functions without Wrapper in the name; however, any can be
 * successfully used.
 */

type KeyBasedStruct struct {
	Id          string `dynamodbav:"id"`
	SecondaryId string `dynamodbav:"secondaryId"`
}

func ddbPutWrapper(item interface{}, conditionExp *string) (*dynamodb.PutItemOutput, error) {
	ddbClient := GetDynamodbClient()
	av, marshalErr := attributevalue.MarshalMap(item)
	if marshalErr != nil {
		logger.Error("Failed to marshal item",
			zap.Any("item", item),
			zap.Error(marshalErr),
		)
		return &dynamodb.PutItemOutput{}, marshalErr
	}

	putItemRes, putItemErr := ddbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           aws.String(config.PrimaryTableName),
		Item:                av,
		ConditionExpression: conditionExp,
	})
	if putItemErr != nil {
		logger.Error("Failed to put item", zap.Error(putItemErr))
		return &dynamodb.PutItemOutput{}, putItemErr
	}

	return putItemRes, nil
}

func ddbGetWrapper(key interface{}, resultItem interface{}) (*dynamodb.GetItemOutput, error) {
	ddbClient := GetDynamodbClient()
	av, marshalErr := attributevalue.MarshalMap(key)
	if marshalErr != nil {
		logger.Error("Failed to marshal key",
			zap.Any("key", key),
			zap.Error(marshalErr),
		)
		return &dynamodb.GetItemOutput{}, marshalErr
	}

	getItemRes, getItemErr := ddbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(config.PrimaryTableName),
		Key:       av,
	})
	if getItemErr != nil {
		logger.Error("Failed to get item", zap.Error(getItemErr))
		return &dynamodb.GetItemOutput{}, getItemErr
	}
	unmarshalErr := attributevalue.UnmarshalMap(getItemRes.Item, resultItem)
	if unmarshalErr != nil {
		logger.Error("Failed to unmarshal item",
			zap.Error(unmarshalErr),
		)
		return &dynamodb.GetItemOutput{}, unmarshalErr
	}

	return getItemRes, nil
}

func ddbQueryWrapper(key string, limit int32, startKey map[string]ddbtypes.AttributeValue) (*dynamodb.QueryOutput, error) {
	ddbClient := GetDynamodbClient()

	keyExpr := expression.Key("id").Equal(expression.Value(key))
	expr, builderErr := expression.NewBuilder().WithKeyCondition(keyExpr).Build()
	if builderErr != nil {
		logger.Error("Failed to build key condition expression",
			zap.Error(builderErr),
		)
		return &dynamodb.QueryOutput{}, builderErr
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(config.PrimaryTableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(limit),
	}

	if len(startKey) != 0 {
		queryInput.ExclusiveStartKey = startKey
	}

	queryRes, queryErr := ddbClient.Query(context.TODO(), queryInput)
	if queryErr != nil {
		logger.Error("Failed query", zap.Error(queryErr))
		return &dynamodb.QueryOutput{}, queryErr
	}

	return queryRes, nil
}

func ddbUpdateWrapper(key interface{}, update expression.UpdateBuilder) (*dynamodb.UpdateItemOutput, error) {
	ddbClient := GetDynamodbClient()
	av, marshalErr := attributevalue.MarshalMap(key)
	if marshalErr != nil {
		logger.Error("Failed to marshal key",
			zap.Any("key", key),
			zap.Error(marshalErr),
		)
		return &dynamodb.UpdateItemOutput{}, marshalErr
	}

	expr, builderErr := expression.NewBuilder().WithUpdate(update).Build()
	if builderErr != nil {
		logger.Error("Failed to build update expression",
			zap.Error(builderErr),
		)
		return &dynamodb.UpdateItemOutput{}, builderErr
	}

	updateItemRes, updateItemErr := ddbClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 aws.String(config.PrimaryTableName),
		Key:                       av,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              ddbtypes.ReturnValueAllNew,
	})
	if updateItemErr != nil {
		logger.Error("Failed to update item", zap.Error(updateItemErr))
		return &dynamodb.UpdateItemOutput{}, updateItemErr
	}

	return updateItemRes, nil
}

func ddbDeleteWrapper(key interface{}) (*dynamodb.DeleteItemOutput, error) {
	ddbClient := GetDynamodbClient()
	av, marshalErr := attributevalue.MarshalMap(key)
	if marshalErr != nil {
		logger.Error("Failed to marshal key",
			zap.Any("key", key),
			zap.Error(marshalErr),
		)
		return &dynamodb.DeleteItemOutput{}, marshalErr
	}

	deleteItemRes, deleteItemErr := ddbClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(config.PrimaryTableName),
		Key:       av,
	})
	if deleteItemErr != nil {
		logger.Error("Failed to delete item", zap.Error(deleteItemErr))
		return &dynamodb.DeleteItemOutput{}, deleteItemErr
	}

	return deleteItemRes, nil
}

func ddbBulkDeleteWrapper(writeReqs []ddbtypes.WriteRequest) (*dynamodb.BatchWriteItemOutput, error) {
	ddbClient := GetDynamodbClient()
	batchDeleteOutput, err := ddbClient.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]ddbtypes.WriteRequest{
			config.PrimaryTableName: writeReqs,
		},
	})
	return batchDeleteOutput, err
}

func ddbOverwrite(item interface{}) (*dynamodb.PutItemOutput, error) {
	putItemRes, putItemErr := ddbPutWrapper(item, nil)

	return putItemRes, putItemErr
}

func ddbPut(item interface{}) (*dynamodb.PutItemOutput, error) {
	putItemRes, putItemErr := ddbPutWrapper(item, aws.String("attribute_not_exists(secondaryId)"))

	return putItemRes, putItemErr
}

func ddbGet(key interface{}, resultItem interface{}) (*dynamodb.GetItemOutput, error) {
	return ddbGetWrapper(key, resultItem)
}

// TODO is there a way to genericize the queries?
// Can't pass []interface{} so each type needs its own function
func ddbQueryEntitys(key string, limit int, nextToken string) ([]types.Entity, string, error) {
	entitys := make([]types.Entity, 0)

	// Make empty map and check in ddb query wrapper if it is empty
	startKey := make(map[string]ddbtypes.AttributeValue)
	if nextToken != "" {
		exclStartString, decErr := base64.StdEncoding.DecodeString(nextToken)
		if decErr != nil {
			logger.Error("Failed to decode nextToken base64",
				zap.Error(decErr),
			)
			return entitys, "", decErr
		}

		entity := &types.DdbPrimaryKey{}
		jsonErr := json.Unmarshal(exclStartString, entity)
		if jsonErr != nil {
			logger.Error("Failed to unmarshal nextToken json",
				zap.Error(jsonErr),
			)
			return entitys, "", jsonErr
		}

		marshalledStartKey, marshalErr := attributevalue.MarshalMap(entity)
		if marshalErr != nil {
			logger.Error("Failed to marshal entity to map[string]AttributeValue",
				zap.Error(marshalErr),
			)
			return entitys, "", marshalErr
		}

		startKey = marshalledStartKey
	}

	queryRes, err := ddbQueryWrapper(key, int32(limit), startKey)
	if err != nil {
		return entitys, "", err
	}

	for _, item := range queryRes.Items {
		ddbEntity := &types.DdbEntityItem{}
		unmarshalErr := attributevalue.UnmarshalMap(item, ddbEntity)
		if unmarshalErr != nil {
			logger.Error("Failed to unmarshal entity from list",
				zap.Error(unmarshalErr),
			)
		}

		entity := normalizeDdbEntity(ddbEntity)

		entitys = append(entitys, entity)
	}

	if len(queryRes.LastEvaluatedKey) != 0 {
		lastEvalKey := &types.DdbPrimaryKey{}
		marshalErr := attributevalue.UnmarshalMap(queryRes.LastEvaluatedKey, lastEvalKey)
		if marshalErr != nil {
			logger.Error("Failed to unmarshal map[string]AttributeValue to entity",
				zap.Error(marshalErr),
			)
			return entitys, "", marshalErr
		}

		lastEvalString, jsonErr := json.Marshal(lastEvalKey)
		if jsonErr != nil {
			logger.Error("Failed to marshal last evaluated key json",
				zap.Error(jsonErr),
			)
			return entitys, "", jsonErr
		}

		lastEvalB64 := base64.StdEncoding.EncodeToString([]byte(lastEvalString))

		return entitys, lastEvalB64, nil
	}

	return entitys, "", nil
}

func ddbUpdate(key interface{}, update expression.UpdateBuilder) (*dynamodb.UpdateItemOutput, error) {
	return ddbUpdateWrapper(key, update)
}

func ddbUpdateAndReturn(key interface{}, update expression.UpdateBuilder, resultItem interface{}) (*dynamodb.UpdateItemOutput, error) {
	updateOutput, err := ddbUpdateWrapper(key, update)
	if err != nil {
		return updateOutput, err
	}

	unmarshalErr := attributevalue.UnmarshalMap(updateOutput.Attributes, resultItem)
	if unmarshalErr != nil {
		logger.Error("Failed to unmarshal item",
			zap.Error(unmarshalErr),
		)
		return updateOutput, unmarshalErr
	}

	return updateOutput, nil
}

func ddbDelete(key interface{}) (*dynamodb.DeleteItemOutput, error) {
	return ddbDeleteWrapper(key)
}
