//go:build integration

package microservice_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"testing"

	redis "github.com/go-redis/redis/v8"
)

func TestRedisClientConnect(t *testing.T) {
	// new redis client
	// NOTE: ต้องตั้ง environment variables ก่อนรัน test:
	//   REDIS_ADDR, REDIS_USERNAME, REDIS_PASSWORD

	addr := os.Getenv("REDIS_ADDR")
	username := os.Getenv("REDIS_USERNAME")
	password := os.Getenv("REDIS_PASSWORD")

	if addr == "" {
		t.Skip("REDIS_ADDR not set, skipping integration test")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       1,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})

	// test connection

	pong, err := client.Ping(context.TODO()).Result()

	if err != nil {

		t.Error(err)

	}

	// return pong if server is online

	fmt.Println(pong)
}
