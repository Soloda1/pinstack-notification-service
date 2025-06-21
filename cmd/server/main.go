package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"os"
	"os/signal"
	"pinstack-notification-service/config"
	notification_grpc "pinstack-notification-service/internal/delivery/grpc"
	"pinstack-notification-service/internal/delivery/kafka/consumer"
	"pinstack-notification-service/internal/logger"
	repository_postgres "pinstack-notification-service/internal/repository/notification/postgres"
	notification_service "pinstack-notification-service/internal/service/notification"
	"syscall"
	"time"
)

func main() {
	cfg := config.MustLoad()
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DbName)
	ctx := context.Background()
	log := logger.New(cfg.Env)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("Failed to parse postgres poolConfig", slog.String("error", err.Error()))
		os.Exit(1)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Error("Failed to create postgres pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	notificationRepo := repository_postgres.NewNotificationRepository(pool, log)

	notificationService := notification_service.NewNotificationService(log, notificationRepo)

	kafkaConsumer, err := consumer.NewNotificationConsumer(cfg.Kafka, log, notificationService)
	if err != nil {
		log.Error("Failed to initialize Kafka consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}

	notificationGRPCApi := notification_grpc.NewNotificationGRPCService(notificationService, log)
	grpcServer := notification_grpc.NewServer(notificationGRPCApi, cfg.GrpcServer.Address, cfg.GrpcServer.Port, log)

	go kafkaConsumer.Start(ctx)

	done := make(chan bool)
	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Error("gRPC server error", slog.String("error", err.Error()))
		}
		done <- true
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down services...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = shutdownCtx
	defer cancel()

	if err := grpcServer.Shutdown(); err != nil {
		log.Error("gRPC server shutdown error", slog.String("error", err.Error()))
	}

	<-done

	kafkaConsumer.Close()

	log.Info("Server exiting")
}
