package main

import (
	"github.com/thomasstep/giphy-livechat-api/internal/adapters"
	"github.com/thomasstep/giphy-livechat-api/internal/types"
)

func logic(entityId string, updates types.EntityUpdates) (*types.Entity, error) {
	updatedEntity, err := adapters.UpdateEntity(entityId, updates, false)
	if err != nil {
		return &types.Entity{}, err
	}

	return updatedEntity, nil
}
