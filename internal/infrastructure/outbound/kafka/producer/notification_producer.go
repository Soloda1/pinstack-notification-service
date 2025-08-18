package producer

import (
	"context"
	"log/slog"
	ports "pinstack-notification-service/internal/domain/ports/output"
	"pinstack-notification-service/internal/infrastructure/config"
	"time"
)

type NotificationProducer struct {
	config  config.KafkaConfig
	log     ports.Logger
	metrics ports.MetricsProvider
}

func NewNotificationProducer(cfg config.KafkaConfig, log ports.Logger, metrics ports.MetricsProvider) *NotificationProducer {
	return &NotificationProducer{
		config:  cfg,
		log:     log,
		metrics: metrics,
	}
}

func (p *NotificationProducer) Send(ctx context.Context, topic string, message []byte) error {
	start := time.Now()
	var success bool
	defer func() {
		p.metrics.IncrementKafkaMessages(topic, "produce", success)
		p.metrics.RecordKafkaMessageDuration(topic, "produce", time.Since(start))
	}()

	p.log.Debug("Sending message to Kafka topic", slog.String("topic", topic))
	// TODO: Реализовать логику отправки в Kafka

	success = true
	return nil
}
