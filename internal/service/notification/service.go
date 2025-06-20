package notification_service

import (
	"pinstack-notification-service/internal/logger"
	notification_repository "pinstack-notification-service/internal/repository/notification"
)

type Service struct {
	notificationRepo notification_repository.NotificationRepository
	log              *logger.Logger
}

func NewNotificationService(log *logger.Logger, notificationRepo notification_repository.NotificationRepository) *Service {
	return &Service{
		log:              log,
		notificationRepo: notificationRepo,
	}
}
