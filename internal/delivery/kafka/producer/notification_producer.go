package producer

import (
	"context"
	"log/slog"
	"pinstack-notification-service/config"
	"pinstack-notification-service/internal/logger"
)

type NotificationProducer struct {
	config config.KafkaConfig
	log    *logger.Logger
}

func NewNotificationProducer(cfg config.KafkaConfig, log *logger.Logger) *NotificationProducer {
	return &NotificationProducer{
		config: cfg,
		log:    log,
	}
}

func (p *NotificationProducer) Send(ctx context.Context, topic string, message []byte) error {
	p.log.Debug("Sending message to Kafka topic", slog.String("topic", topic))
	// TODO: Реализовать логику отправки в Kafka
	return nil
}
