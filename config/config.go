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
}

type GrpcServerConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type EventTypesConfig struct {
	FollowCreated string `yaml:"follow_created"`
	FollowDeleted string `yaml:"follow_deleted"`
}

type Config struct {
	Env        string           `yaml:"env"`
	GrpcServer GrpcServerConfig `yaml:"grpc_server"`
	Kafka      KafkaConfig      `yaml:"kafka"`
	Database   Database         `yaml:"database"`
	EventTypes EventTypesConfig `yaml:"event_types"`
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
		},
		Database: Database{
			Username:       viper.GetString("database.username"),
			Password:       viper.GetString("database.password"),
			Host:           viper.GetString("database.host"),
			Port:           viper.GetString("database.port"),
			DbName:         viper.GetString("database.db_name"),
			MigrationsPath: viper.GetString("database.migrations_path"),
		},
		EventTypes: EventTypesConfig{
			FollowCreated: viper.GetString("event_types.follow_created"),
			FollowDeleted: viper.GetString("event_types.follow_deleted"),
		},
	}

	return config
}
