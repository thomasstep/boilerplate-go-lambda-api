package adapters

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	// "go.uber.org/zap"

	"github.com/thomasstep/giphy-livechat-api/internal/common"
	"github.com/thomasstep/giphy-livechat-api/internal/types"
)

func CreateEntity(entity types.Entity) error {
	entityItem := types.DdbEntityItem{
		Entity:        entity,
		Id:          entity.Id,
		SecondaryId: config.EntitySortKey,
		CreatedTime: common.GetIsoString(),
		UpdatedTime: common.GetIsoString(),
	}

	_, entityPutItemErr := ddbPut(entityItem)
	if entityPutItemErr != nil {
		return entityPutItemErr
	}

	return nil
}

func normalizeDdbEntity(ddb *types.DdbEntityItem) types.Entity {
	// Reconstruct the entity based on user entity items (see top comment or README)
	entity := ddb.Entity
	entity.Id = ddb.SecondaryId
	return entity
}

func ReadEntity(entityId string) (*types.Entity, error) {
	key := &KeyBasedStruct{
		Id:          entityId,
		SecondaryId: config.EntitySortKey,
	}

	result := &types.DdbEntityItem{}
	_, getItemErr := ddbGet(key, result)
	if getItemErr != nil {
		return &types.Entity{}, getItemErr
	}

	entity := result.Entity
	entity.Id = result.SecondaryId

	return &entity, nil
}

func UpdateEntity(entityId string, updated types.EntityUpdates, asOwner bool) (*types.Entity, error) {
	entityKey := &KeyBasedStruct{
		Id:          entityId,
		SecondaryId: config.EntitySortKey,
	}

	var updates expression.UpdateBuilder
	if updated.Name != "" {
		updates = updates.Set(
			expression.Name("name"),
			expression.Value(updated.Name),
		)
	} else {
		// Nothing to do
		return &types.Entity{}, nil
	}

	result := &types.DdbEntityItem{}
	_, updateItemErr := ddbUpdateAndReturn(entityKey, updates, result)
	if updateItemErr != nil {
		return &types.Entity{}, updateItemErr
	}

	updatedEntity := normalizeDdbEntity(result)

	return &updatedEntity, nil
}

// Only delete the main entity if the owner is performing the action
func DeleteEntity(entityId string, asOwner bool) error {
	entityKey := &KeyBasedStruct{
		Id:          entityId,
		SecondaryId: config.EntitySortKey,
	}

	_, deleteUserItemErr := ddbDelete(entityKey)

	if deleteUserItemErr != nil {
		return deleteUserItemErr
	}

	return nil
}
