package config

import (
	"sync"

	"github.com/thomasstep/giphy-livechat-api/internal/common"
)

type ConfigStruct struct {
	Region string

	// Authorization service
	// JwksUrl string
	// JwkId   string
	// AuthUrl string

	// Database related
	PrimaryTableName string
	Limit            int
	EntitySortKey  string

	// SNS related
	PrimaryTopicArn          string
}

var Config *ConfigStruct
var onceConfig sync.Once

func GetConfig() *ConfigStruct {
	onceConfig.Do(func() {
		Config = &ConfigStruct{
			Region:                   common.GetEnv("AWS_REGION", "us-east-1"),
			// JwksUrl:                  common.GetEnv("JWKS_URL", ""),
			// JwkId:                    common.GetEnv("JWK_ID", ""),
			// AuthUrl:                  common.GetEnv("AUTH_URL", ""),
			PrimaryTableName:         common.GetEnv("PRIMARY_TABLE_NAME", ""),
			Limit:                    20, // BatchWrite on DDB has limit of 25
			EntitySortKey:          "entity",
			PrimaryTopicArn:          common.GetEnv("PRIMARY_SNS_TOPIC_ARN", ""),
		}
	})
	return Config
}
