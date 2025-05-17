package config

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
)

func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS"),
		Password: "",
		DB:       0,
	})

	// Kiểm tra kết nối
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Fail to connent with Redis: %v", err)
	}

	return client
}
