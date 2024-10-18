package redis_factory

import (
	"context"
	"gateway_api/app/utils/yml_config"
	"gateway_api/app/utils/yml_config/ymlconfig_interf"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var configYml ymlconfig_interf.YmlConfigInterf

func init() {
	configYml = yml_config.CreateYamlFactory()
	Redis()
}

// init redis
func Redis() {
	client := redis.NewClient(&redis.Options{
		Addr:           configYml.GetString("Redis.Host") + ":" + configYml.GetString("Redis.Port"),
		Password:       configYml.GetString("Redis.Auth"), // 填入自己的Redis密码默认没有
		DB:             configYml.GetInt("Redis.IndexDb"),
		MaxRetries:     configYml.GetInt("Redis.ConnFailRetryTimes"),
		MaxIdleConns:   configYml.GetInt("Redis.MaxIdle"), //最大空闲数
		MaxActiveConns: configYml.GetInt("Redis.MaxActive"),
	})

	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		panic("can't connect redis")
	}

	RedisClient = client
}
