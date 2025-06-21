package consumer

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log/slog"
	"pinstack-notification-service/config"
	"pinstack-notification-service/internal/logger"
)

type NotificationConsumer struct {
	config   config.KafkaConfig
	log      *logger.Logger
	consumer *kafka.Consumer
}

func NewNotificationConsumer(cfg config.KafkaConfig, log *logger.Logger) (*NotificationConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       cfg.Brokers,
		"group.id":                cfg.ConsumerGroupID,
		"auto.offset.reset":       cfg.AutoOffsetReset,
		"enable.auto.commit":      cfg.EnableAutoCommit,
		"auto.commit.interval.ms": cfg.AutoCommitIntervalMs,
		"session.timeout.ms":      cfg.SessionTimeoutMs,
		"max.poll.records":        cfg.MaxPollRecords,
		"max.poll.interval.ms":    cfg.MaxPollIntervalMs,
	})

	if err != nil {
		log.Error("Failed to create Kafka consumer", slog.String("error", err.Error()))
		return nil, err
	}

	return &NotificationConsumer{
		config:   cfg,
		log:      log,
		consumer: c,
	}, nil
}

func (c *NotificationConsumer) Start(ctx context.Context) {
	c.log.Info("Starting Kafka consumer")
	// TODO: Реализовать логику чтения сообщений из Kafka
}
