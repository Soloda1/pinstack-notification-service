package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/soloda1/pinstack-proto-definitions/events"
	"log/slog"
	"pinstack-notification-service/config"
	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	"pinstack-notification-service/internal/model"
	notification_service "pinstack-notification-service/internal/service/notification"
	"time"
)

type NotificationConsumer struct {
	config              config.KafkaConfig
	log                 *logger.Logger
	consumer            *kafka.Consumer
	notificationService notification_service.NotificationService
}

func NewNotificationConsumer(cfg config.KafkaConfig, log *logger.Logger, notificationSvc notification_service.NotificationService) (*NotificationConsumer, error) {
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
		config:              cfg,
		log:                 log,
		consumer:            c,
		notificationService: notificationSvc,
	}, nil
}

func (c *NotificationConsumer) Start(ctx context.Context) {
	c.log.Info("Starting Kafka consumer", slog.String("topic", c.config.RelationTopic))

	err := c.consumer.SubscribeTopics([]string{c.config.RelationTopic}, nil)
	if err != nil {
		c.log.Error("Failed to subscribe to topic",
			slog.String("topic", c.config.RelationTopic),
			slog.String("error", err.Error()))
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.log.Error("Recovered from panic in Kafka consumer", slog.Any("panic", r))
			}
		}()

		for {
			select {
			case <-ctx.Done():
				c.log.Info("Stopping Kafka consumer", slog.String("reason", "context done"))
				c.Close()
				return
			default:
				msg, err := c.consumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					if !errors.Is(err, kafka.NewError(kafka.ErrTimedOut, "", false)) {
						c.log.Error("Error reading message from Kafka", slog.String("error", err.Error()))
					}
					continue
				}

				if msg == nil {
					continue
				}

				if err := c.processMessage(ctx, msg); err != nil {
					c.log.Error("Failed to process message",
						slog.String("topic", *msg.TopicPartition.Topic),
						slog.Int("partition", int(msg.TopicPartition.Partition)),
						slog.Int64("offset", int64(msg.TopicPartition.Offset)),
						slog.String("key", string(msg.Key)),
						slog.String("error", err.Error()))
				} else {
					c.log.Info("Message processed successfully",
						slog.String("topic", *msg.TopicPartition.Topic),
						slog.Int("partition", int(msg.TopicPartition.Partition)),
						slog.Int64("offset", int64(msg.TopicPartition.Offset)),
						slog.String("key", string(msg.Key)))

					if !c.config.EnableAutoCommit {
						if _, err := c.consumer.Commit(); err != nil {
							c.log.Error("Failed to commit offset",
								slog.String("topic", *msg.TopicPartition.Topic),
								slog.Int("partition", int(msg.TopicPartition.Partition)),
								slog.Int64("offset", int64(msg.TopicPartition.Offset)),
								slog.String("error", err.Error()))
						}
					}
				}
			}
		}
	}()
}

func (c *NotificationConsumer) processMessage(ctx context.Context, msg *kafka.Message) error {
	if msg == nil || msg.Value == nil {
		return custom_errors.ErrInvalidInput
	}

	var event struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.log.Error("Failed to unmarshal Kafka message",
			slog.String("message", string(msg.Value)),
			slog.String("error", err.Error()))
		return custom_errors.ErrInvalidInput
	}

	c.log.Info("Received event from Kafka",
		slog.String("event_type", event.EventType),
		slog.String("payload", string(event.Payload)))

	switch event.EventType {
	case string(events.EventTypeFollowCreated):
		return c.handleFollowCreated(ctx, event.Payload)
	default:
		c.log.Warn("Unknown event type", slog.String("event_type", event.EventType))
		return custom_errors.ErrInvalidInput
	}
}

func (c *NotificationConsumer) handleFollowCreated(ctx context.Context, payload json.RawMessage) error {
	var followEvent events.FollowCreatedPayload
	if err := json.Unmarshal(payload, &followEvent); err != nil {
		c.log.Error("Failed to unmarshal follow created event",
			slog.String("payload", string(payload)),
			slog.String("error", err.Error()))
		return custom_errors.ErrInvalidInput
	}

	if followEvent.FolloweeID <= 0 || followEvent.FollowerID <= 0 {
		c.log.Error("Invalid follow event data",
			slog.Int64("follower_id", followEvent.FollowerID),
			slog.Int64("followee_id", followEvent.FolloweeID))
		return custom_errors.ErrInvalidInput
	}

	notification := &model.Notification{
		UserID:    followEvent.FolloweeID,
		Type:      events.EventTypeFollowCreated,
		IsRead:    false,
		CreatedAt: time.Now(),
		Payload:   payload,
	}

	c.log.Info("Created follow notification",
		slog.Int64("user_id", notification.UserID),
		slog.String("type", string(notification.Type)),
		slog.String("payload", string(notification.Payload)))

	err := c.notificationService.SaveNotification(ctx, notification)
	if err != nil {
		c.log.Error("Failed to save notification", slog.String("error", err.Error()))
		return err
	}

	c.log.Info("Notification saved successfully", slog.Int64("user_id", notification.UserID))
	return nil
}

func (c *NotificationConsumer) Close() {
	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			c.log.Error("Failed to close Kafka consumer", slog.String("error", err.Error()))
		} else {
			c.log.Info("Kafka consumer closed successfully")
		}
	}
}
