package main

import (
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// redisConnect is used to manage the connection to redis and gracefully exit if the connection fails
func redisConnect(addr string, password string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Err(err).
			Str("func", "redisConnect").
			Msg("Unable to connect to redis. Exiting!")
		os.Exit(1)
	}

	return rdb
}
