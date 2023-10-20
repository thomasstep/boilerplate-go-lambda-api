package main

import (
	"go.uber.org/zap"

	configMod "github.com/thomasstep/giphy-livechat-api/internal/common/config"
)

var logger *zap.Logger
var config *configMod.ConfigStruct

func init() {
	logger = zap.NewExample()
	defer logger.Sync()

	config = configMod.GetConfig()
}