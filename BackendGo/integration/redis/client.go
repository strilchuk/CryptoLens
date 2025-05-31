package redis

import (
	"CryptoLens_Backend/env"
	"CryptoLens_Backend/logger"
	"context"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init() error {
	Client = redis.NewClient(&redis.Options{
		Addr:     env.GetRedisHost() + ":" + env.GetRedisPortLocal(),
		Password: env.GetRedisPassword(),
		DB:       0,
	})

	_, err := Client.Ping(context.Background()).Result()
	if err != nil {
		logger.LogError("Failed to connect to Redis: %v", err)
		return err
	}
	logger.LogInfo("Successfully connected to Redis")
	return nil
} 