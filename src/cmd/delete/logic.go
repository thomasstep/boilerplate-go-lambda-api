package main

import (
	"github.com/thomasstep/giphy-livechat-api/internal/adapters"
)

func logic(entityId string) error {
	return adapters.DeleteEntity(entityId, true)
}
