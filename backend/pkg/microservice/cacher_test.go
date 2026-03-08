package microservice_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"testing"

	redis "github.com/go-redis/redis/v8"
)

func TestRedisClientConnect(t *testing.T) {
	// new redis client

	client := redis.NewClient(&redis.Options{

		Addr:     "db-redis-sgp1-13586-do-user-11230406-0.b.db.ondigitalocean.com:25061",
		Username: "default",
		Password: "AVNS_dhilsnrZwup9NhqL8GQ",

		DB: 1,
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
