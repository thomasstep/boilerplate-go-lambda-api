package main

import (
	"errors"

	"github.com/thomasstep/giphy-livechat-api/internal/adapters"
	"github.com/thomasstep/giphy-livechat-api/internal/types"
)

func logic(entityId string) (*types.Entity, error) {
	entity, err := adapters.ReadEntity(entityId)
	if err != nil {
		return &types.Entity{}, err
	}

	if entity.Id == "" {
		return &types.Entity{}, &types.MissingResourceError{
			Err: errors.New("Could not find entity."),
		}
	}
	return entity, nil
}
