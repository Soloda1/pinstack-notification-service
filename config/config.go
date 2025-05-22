package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	GroupID string   `yaml:"group_id"`
	Topics  []string `yaml:"topics"`
}

type Config struct {
	Env      string
	Kafka    KafkaConfig `yaml:"kafka"`
	Database Database    `yaml:"database"`
}

type Database struct {
	Username       string
	Password       string
	Host           string
	Port           string
	DbName         string
	MigrationsPath string
}

func MustLoad() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.SetDefault("env", "dev")

	viper.SetDefault("kafka.group_id", "notification-group")
	viper.SetDefault("kafka.brokers", []string{"kafka:9092"})
	viper.SetDefault("kafka.topics", []string{"user-notifications", "post-notifications"})

	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "admin")
	viper.SetDefault("database.host", "notification-db")
	viper.SetDefault("database.port", "5435")
	viper.SetDefault("database.db_name", "notificationservice")
	viper.SetDefault("database.migrations_path", "migrations")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
		os.Exit(1)
	}

	config := &Config{
		Env: viper.GetString("env"),
		Kafka: KafkaConfig{
			Brokers: viper.GetStringSlice("kafka.brokers"),
			GroupID: viper.GetString("kafka.group_id"),
			Topics:  viper.GetStringSlice("kafka.topics"),
		},
		Database: Database{
			Username:       viper.GetString("database.username"),
			Password:       viper.GetString("database.password"),
			Host:           viper.GetString("database.host"),
			Port:           viper.GetString("database.port"),
			DbName:         viper.GetString("database.db_name"),
			MigrationsPath: viper.GetString("database.migrations_path"),
		},
	}

	return config
}
