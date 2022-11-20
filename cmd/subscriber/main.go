package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"user-microservice/config"
	"user-microservice/internal/models"
	userPubSub "user-microservice/internal/users/pubsub"
	redisDB "user-microservice/pkg/db/redis"

	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("Redis listener")

	filepath := os.Getenv("CONFIG_FILE")
	cfg, err := config.GetConfigFromFile(filepath)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	redisClient := redisDB.MewRedisDatabase(cfg.Redis)

	subscriber := redisClient.Subscribe(ctx, userPubSub.GetAllUsersTopics()...)

	var payload interface{}
	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			logrus.Errorf("Error receiving message: %s", err)
		} else {

			if msg.Channel == userPubSub.TopicUserDeletion {
				payload = msg.Payload
			} else {
				var user models.User
				if err := json.Unmarshal([]byte(msg.Payload), &user); err != nil {
					logrus.Errorf("Error unmarshaling payload: %s", err)
				} else {
					payload = user
				}
			}

			fmt.Println("Received message from " + msg.Channel + " channel.")
			fmt.Printf("Payload: %+v\n", payload)
		}
	}

}
