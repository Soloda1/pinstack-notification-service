package producer

import (
	"context"
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
	p.log.Debug("Sending message to topic")
	// TODO: Реализовать логику отправки в Kafka
	return nil
}
