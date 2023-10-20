package adapters

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"

	"github.com/thomasstep/giphy-livechat-api/internal/types"
)

type EventActionEvent struct {
	Entity    *types.Entity       `json:"entity"`
	Updates types.EntityUpdates `json:"updates"`
}

func EmitEventAction(entity *types.Entity, updates types.EntityUpdates) error {
	message := &EventActionEvent{
		Entity:    entity,
		Updates: updates,
	}
	messageAttributes := map[string]snstypes.MessageAttributeValue{
		"operation": snstypes.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String("entityUpdated"),
		},
	}

	_, publishErr := snsPublish(message, messageAttributes)
	if publishErr != nil {
		return publishErr
	}

	return nil
}
