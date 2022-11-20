package main

import (
	"context"
	"encoding/json"
	"fmt"
	"user-microservice/internal/models"
	userPubSub "user-microservice/internal/users/pubsub"
	redisDB "user-microservice/pkg/db/redis"
)

func main() {
	fmt.Println("Redis listener")
	ctx := context.Background()
	redisClient := redisDB.MewRedisDatabase()

	subscriber := redisClient.Subscribe(ctx, userPubSub.GetAllUsersTopics()...)

	var user models.User

	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}

		if err := json.Unmarshal([]byte(msg.Payload), &user); err != nil {
			panic(err)
		}

		fmt.Println("Received message from " + msg.Channel + " channel.")
		fmt.Printf("%+v\n", user)
	}

}
