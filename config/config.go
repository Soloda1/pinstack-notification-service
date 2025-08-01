package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type KafkaConfig struct {
	Brokers               string `yaml:"brokers"`
	Acks                  string `yaml:"acks"`
	Retries               int    `yaml:"retries"`
	RetryBackoffMs        int    `yaml:"retry_backoff_ms"`
	DeliveryTimeoutMs     int    `yaml:"delivery_timeout_ms"`
	QueueBufferingMaxMsgs int    `yaml:"queue_buffering_max_messages"`
	QueueBufferingMaxMs   int    `yaml:"queue_buffering_max_ms"`
	CompressionType       string `yaml:"compression_type"`
	BatchSize             int    `yaml:"batch_size"`
	LingerMs              int    `yaml:"linger_ms"`
	RelationTopic         string `yaml:"relation_topic"`
	ConsumerGroupID       string `yaml:"consumer_group_id"`
	AutoOffsetReset       string `yaml:"auto_offset_reset"`
	EnableAutoCommit      bool   `yaml:"enable_auto_commit"`
	AutoCommitIntervalMs  int    `yaml:"auto_commit_interval_ms"`
	SessionTimeoutMs      int    `yaml:"session_timeout_ms"`
	MaxPollIntervalMs     int    `yaml:"max_poll_interval_ms"`
}

type GrpcServerConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type EventTypesConfig struct {
	FollowCreated string `yaml:"follow_created"`
	FollowDeleted string `yaml:"follow_deleted"`
}

type PrometheusConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type Config struct {
	Env         string           `yaml:"env"`
	GrpcServer  GrpcServerConfig `yaml:"grpc_server"`
	Kafka       KafkaConfig      `yaml:"kafka"`
	Database    Database         `yaml:"database"`
	EventTypes  EventTypesConfig `yaml:"event_types"`
	Prometheus  PrometheusConfig `yaml:"prometheus"`
	UserService UserService      `yaml:"user_service"`
}

type UserService struct {
	Address string
	Port    int
}

type Database struct {
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	DbName         string `yaml:"db_name"`
	MigrationsPath string `yaml:"migrations_path"`
}

func MustLoad() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.SetDefault("env", "dev")

	// GRPC Server defaults
	viper.SetDefault("grpc_server.address", "0.0.0.0")
	viper.SetDefault("grpc_server.port", 50055)

	// Kafka defaults
	viper.SetDefault("kafka.brokers", "kafka1:9092,kafka2:9092,kafka3:9092")
	viper.SetDefault("kafka.acks", "all")
	viper.SetDefault("kafka.retries", 3)
	viper.SetDefault("kafka.retry_backoff_ms", 500)
	viper.SetDefault("kafka.delivery_timeout_ms", 5000)
	viper.SetDefault("kafka.queue_buffering_max_messages", 100000)
	viper.SetDefault("kafka.queue_buffering_max_ms", 5)
	viper.SetDefault("kafka.compression_type", "snappy")
	viper.SetDefault("kafka.batch_size", 16384)
	viper.SetDefault("kafka.linger_ms", 5)
	viper.SetDefault("kafka.relation_topic", "relation-events")

	// Kafka consumer defaults
	viper.SetDefault("kafka.consumer_group_id", "notification-service")
	viper.SetDefault("kafka.auto_offset_reset", "earliest")
	viper.SetDefault("kafka.enable_auto_commit", true)
	viper.SetDefault("kafka.auto_commit_interval_ms", 5000)
	viper.SetDefault("kafka.session_timeout_ms", 10000)
	viper.SetDefault("kafka.max_poll_interval_ms", 300000)

	// Event Types defaults
	viper.SetDefault("event_types.follow_created", "follow_created")
	viper.SetDefault("event_types.follow_deleted", "follow_deleted")

	// Database defaults
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "admin")
	viper.SetDefault("database.host", "notification-db")
	viper.SetDefault("database.port", "5436")
	viper.SetDefault("database.db_name", "notificationservice")
	viper.SetDefault("database.migrations_path", "./migrations")

	// User service defaults
	viper.SetDefault("user_service.address", "user-service")
	viper.SetDefault("user_service.port", 50051)

	// Prometheus defaults
	viper.SetDefault("prometheus.address", "0.0.0.0")
	viper.SetDefault("prometheus.port", 9105)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
		os.Exit(1)
	}

	config := &Config{
		Env: viper.GetString("env"),
		GrpcServer: GrpcServerConfig{
			Address: viper.GetString("grpc_server.address"),
			Port:    viper.GetInt("grpc_server.port"),
		},
		Kafka: KafkaConfig{
			Brokers:               viper.GetString("kafka.brokers"),
			Acks:                  viper.GetString("kafka.acks"),
			Retries:               viper.GetInt("kafka.retries"),
			RetryBackoffMs:        viper.GetInt("kafka.retry_backoff_ms"),
			DeliveryTimeoutMs:     viper.GetInt("kafka.delivery_timeout_ms"),
			QueueBufferingMaxMsgs: viper.GetInt("kafka.queue_buffering_max_messages"),
			QueueBufferingMaxMs:   viper.GetInt("kafka.queue_buffering_max_ms"),
			CompressionType:       viper.GetString("kafka.compression_type"),
			BatchSize:             viper.GetInt("kafka.batch_size"),
			LingerMs:              viper.GetInt("kafka.linger_ms"),
			RelationTopic:         viper.GetString("kafka.relation_topic"),

			ConsumerGroupID:      viper.GetString("kafka.consumer_group_id"),
			AutoOffsetReset:      viper.GetString("kafka.auto_offset_reset"),
			EnableAutoCommit:     viper.GetBool("kafka.enable_auto_commit"),
			AutoCommitIntervalMs: viper.GetInt("kafka.auto_commit_interval_ms"),
			SessionTimeoutMs:     viper.GetInt("kafka.session_timeout_ms"),
			MaxPollIntervalMs:    viper.GetInt("kafka.max_poll_interval_ms"),
		},
		Database: Database{
			Username:       viper.GetString("database.username"),
			Password:       viper.GetString("database.password"),
			Host:           viper.GetString("database.host"),
			Port:           viper.GetString("database.port"),
			DbName:         viper.GetString("database.db_name"),
			MigrationsPath: viper.GetString("database.migrations_path"),
		},
		UserService: UserService{
			Address: viper.GetString("user_service.address"),
			Port:    viper.GetInt("user_service.port"),
		},
		EventTypes: EventTypesConfig{
			FollowCreated: viper.GetString("event_types.follow_created"),
			FollowDeleted: viper.GetString("event_types.follow_deleted"),
		},
		Prometheus: PrometheusConfig{
			Address: viper.GetString("prometheus.address"),
			Port:    viper.GetInt("prometheus.port"),
		},
	}

	return config
}
