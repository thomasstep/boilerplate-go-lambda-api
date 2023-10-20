package main

import (
	"github.com/thomasstep/giphy-livechat-api/internal/adapters"
	"github.com/thomasstep/giphy-livechat-api/internal/common"
	"github.com/thomasstep/giphy-livechat-api/internal/types"
)

func logic(name string) (*types.Entity, error) {
	entityId := common.GenerateToken()
	entity := types.Entity{
		Id:            entityId,
		Name:          name,
	}
	err := adapters.CreateEntity(entity)
	return &entity, err
}
