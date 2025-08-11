package producer

import (
	"context"
	"log/slog"
	ports "pinstack-notification-service/internal/domain/ports/output"
	"pinstack-notification-service/internal/infrastructure/config"
)

type NotificationProducer struct {
	config config.KafkaConfig
	log    ports.Logger
}

func NewNotificationProducer(cfg config.KafkaConfig, log ports.Logger) *NotificationProducer {
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
