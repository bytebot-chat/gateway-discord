package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/bytebot-chat/gateway-discord/model"
	"github.com/go-redis/redis/v8"
	"github.com/satori/go.uuid"
)

var addr = flag.String("redis", "localhost:6379", "Redis server address")
var inbound = flag.String("inbound", "discord-inbound", "Pubsub queue to listen for new messages")
var outbound = flag.String("outbound", "discord", "Pubsub queue for sending messages outbound")

func main() {
	flag.Parse()
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: *addr,
		DB:   0,
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		time.Sleep(3 * time.Second)
		err := rdb.Ping(ctx).Err()
		if err != nil {
			panic(err)
		}
	}

	topic := rdb.Subscribe(ctx, *inbound)
	channel := topic.Channel()
	for msg := range channel {
		m := &model.Message{}
		err := m.Unmarshal([]byte(msg.Payload))
		if err != nil {
			fmt.Println(err)
		}
		if m.Content == "ping" {
			reply(ctx, *m, rdb)
		}
	}
}

func reply(ctx context.Context, m model.Message, rdb *redis.Client) {
	metadata := model.Metadata{
		Dest:   m.Metadata.Source,
		Source: "discord-pingpong",
		ID:     uuid.Must(uuid.NewV4(), *new(error)),
	}
	stringMsg, _ := m.MarshalReply(metadata, m.ChannelID, "pong")
	rdb.Publish(ctx, *outbound, stringMsg)
	return
}
