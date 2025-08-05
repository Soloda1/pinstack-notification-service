package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"pinstack-notification-service/config"
	user_client "pinstack-notification-service/internal/clients/user"
	notification_grpc "pinstack-notification-service/internal/delivery/grpc"
	"pinstack-notification-service/internal/delivery/kafka/consumer"
	"pinstack-notification-service/internal/logger"
	repository_postgres "pinstack-notification-service/internal/repository/notification/postgres"
	notification_service "pinstack-notification-service/internal/service/notification"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	userServiceConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d", cfg.UserService.Address, cfg.UserService.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("Failed to connect to user service", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func(userServiceConn *grpc.ClientConn) {
		err := userServiceConn.Close()
		if err != nil {
			log.Error("Failed to close user service connection", slog.String("error", err.Error()))
		}
	}(userServiceConn)

	userClient := user_client.NewUserClient(userServiceConn, log)

	notificationRepo := repository_postgres.NewNotificationRepository(pool, log)

	notificationService := notification_service.NewNotificationService(log, notificationRepo, userClient)

	kafkaConsumer, err := consumer.NewNotificationConsumer(cfg.Kafka, log, notificationService)
	if err != nil {
		log.Error("Failed to initialize Kafka consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}

	notificationGRPCApi := notification_grpc.NewNotificationGRPCService(notificationService, log)
	grpcServer := notification_grpc.NewServer(notificationGRPCApi, cfg.GrpcServer.Address, cfg.GrpcServer.Port, log)

	metricsAddr := fmt.Sprintf("%s:%d", cfg.Prometheus.Address, cfg.Prometheus.Port)
	metricsServer := &http.Server{
		Addr:    metricsAddr,
		Handler: nil,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	done := make(chan bool, 1)
	metricsDone := make(chan bool, 1)
	kafkaShutdownDone := make(chan bool, 1)

	go kafkaConsumer.Start(ctx)

	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Error("gRPC server error", slog.String("error", err.Error()))
		}
		done <- true
	}()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Info("Starting Prometheus metrics server", slog.String("address", metricsAddr))
		if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Prometheus metrics server error", slog.String("error", err.Error()))
		}
		metricsDone <- true
	}()

	<-quit
	log.Info("Shutting down services...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	go func() {
		kafkaConsumer.Close()
		kafkaShutdownDone <- true
	}()

	if err := grpcServer.Shutdown(); err != nil {
		log.Error("gRPC server shutdown error", slog.String("error", err.Error()))
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Error("Metrics server shutdown error", slog.String("error", err.Error()))
	}

	select {
	case <-done:
		log.Info("gRPC server shutdown complete")
	case <-shutdownCtx.Done():
		log.Error("gRPC server shutdown timeout exceeded", slog.String("error", shutdownCtx.Err().Error()))
	}

	select {
	case <-metricsDone:
		log.Info("Metrics server shutdown complete")
	case <-shutdownCtx.Done():
		log.Error("Metrics server shutdown timeout exceeded", slog.String("error", shutdownCtx.Err().Error()))
	}

	select {
	case <-kafkaShutdownDone:
		log.Info("Kafka consumer shutdown complete")
	case <-shutdownCtx.Done():
		log.Error("Kafka consumer shutdown timeout exceeded", slog.String("error", shutdownCtx.Err().Error()))
	}

	log.Info("Server exiting")
}
