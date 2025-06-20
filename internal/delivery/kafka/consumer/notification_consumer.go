package consumer

import (
	"context"
	"pinstack-notification-service/config"
	"pinstack-notification-service/internal/logger"
)

type NotificationConsumer struct {
	config config.KafkaConfig
	log    *logger.Logger
}

func NewNotificationConsumer(cfg config.KafkaConfig, log *logger.Logger) *NotificationConsumer {
	return &NotificationConsumer{
		config: cfg,
		log:    log,
	}
}

func (c *NotificationConsumer) Start(ctx context.Context) {
	c.log.Info("Starting Kafka consumer")
	// TODO: Реализовать логику чтения сообщений из Kafka
}
