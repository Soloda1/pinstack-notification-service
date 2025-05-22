package main

import (
	"context"
	"pinstack-notification-service/config"
	"pinstack-notification-service/internal/delivery/kafka/consumer"
	"pinstack-notification-service/internal/logger"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()
	log := logger.New(cfg.Env)

	kafkaConsumer := consumer.NewNotificationConsumer(cfg.Kafka, log)
	go kafkaConsumer.Start(ctx)

	select {}
}
