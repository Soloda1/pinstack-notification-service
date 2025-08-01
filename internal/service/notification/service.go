package notification_service

import (
	"context"
	"errors"
	"log/slog"
	user_client "pinstack-notification-service/internal/clients/user"
	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	"pinstack-notification-service/internal/model"
	notification_repository "pinstack-notification-service/internal/repository/notification"
	"time"
)

type Service struct {
	notificationRepo notification_repository.NotificationRepository
	userClient       user_client.Client
	log              *logger.Logger
}

func NewNotificationService(log *logger.Logger, notificationRepo notification_repository.NotificationRepository, userClient user_client.Client) *Service {
	return &Service{
		log:              log,
		notificationRepo: notificationRepo,
		userClient:       userClient,
	}
}

func (s *Service) SaveNotification(ctx context.Context, notification *model.Notification) (int64, error) {
	if notification == nil {
		s.log.Error("Notification is nil")
		return 0, custom_errors.ErrInvalidInput
	}

	if notification.UserID <= 0 {
		s.log.Error("Invalid user ID in notification", slog.Int64("user_id", notification.UserID))
		return 0, custom_errors.ErrInvalidInput
	}

	_, err := s.userClient.GetUser(ctx, notification.UserID)
	if err != nil {
		s.log.Error("Failed to get user", slog.Int64("user_id", notification.UserID))
		switch {
		case errors.Is(err, custom_errors.ErrUserNotFound):
			s.log.Debug("User not found in save notification", slog.Int64("user_id", notification.UserID), slog.String("error", err.Error()))
			return 0, custom_errors.ErrUserNotFound
		default:
			s.log.Error("Failed to get user", slog.Int64("user_id", notification.UserID))
			return 0, err
		}
	}

	if notification.Type == "" {
		s.log.Error("Empty notification type", slog.Int64("user_id", notification.UserID))
		return 0, custom_errors.ErrInvalidInput
	}

	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}

	notification.IsRead = false

	s.log.Info("Sending notification",
		slog.Int64("user_id", notification.UserID),
		slog.String("type", string(notification.Type)),
	)

	notificationID, err := s.notificationRepo.Create(ctx, notification)
	if err != nil {
		s.log.Error("Failed to send notification",
			slog.Int64("user_id", notification.UserID),
			slog.String("type", string(notification.Type)),
			slog.String("error", err.Error()),
		)
		return 0, err
	}

	s.log.Info("Notification sent successfully",
		slog.Int64("notification_id", notificationID),
		slog.Int64("user_id", notification.UserID),
		slog.String("type", string(notification.Type)),
	)

	return notificationID, nil
}

func (s *Service) GetNotificationDetails(ctx context.Context, id int64) (*model.Notification, error) {
	if id <= 0 {
		s.log.Error("Invalid notification ID", slog.Int64("id", id))
		return nil, custom_errors.ErrInvalidInput
	}

	s.log.Info("Requesting notification details", slog.Int64("id", id))

	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, custom_errors.ErrNotificationNotFound) {
			s.log.Debug("Notification not found", slog.Int64("id", id))
			return nil, custom_errors.ErrNotificationNotFound
		}

		s.log.Error("Failed to retrieve notification details",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.log.Info("Notification details retrieved",
		slog.Int64("id", notification.ID),
		slog.Int64("user_id", notification.UserID),
		slog.String("type", string(notification.Type)),
	)

	return notification, nil
}

func (s *Service) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	if userID <= 0 {
		s.log.Error("Invalid user ID", slog.Int64("user_id", userID))
		return 0, custom_errors.ErrInvalidInput
	}

	s.log.Info("Retrieving unread notification count", slog.Int64("user_id", userID))

	count, err := s.notificationRepo.CountUnread(ctx, userID)
	if err != nil {
		s.log.Error("Failed to retrieve unread notification count",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return 0, err
	}

	s.log.Info("Unread notification count retrieved",
		slog.Int64("user_id", userID),
		slog.Int("count", count),
	)

	return count, nil
}

func (s *Service) GetUserNotificationFeed(ctx context.Context, userID int64, limit, page int) ([]*model.Notification, error) {
	if userID <= 0 {
		s.log.Error("Invalid user ID", slog.Int64("user_id", userID))
		return nil, custom_errors.ErrInvalidInput
	}

	if limit <= 0 {
		s.log.Debug("Using default limit for notifications feed", slog.Int("limit", limit))
		limit = 10
	}

	if page <= 0 {
		s.log.Debug("Using first page for notifications feed", slog.Int("page", page))
		page = 1
	}

	offset := (page - 1) * limit

	s.log.Info("Retrieving user notification feed",
		slog.Int64("user_id", userID),
		slog.Int("limit", limit),
		slog.Int("page", page),
		slog.Int("offset", offset),
	)

	notifications, err := s.notificationRepo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		s.log.Error("Failed to retrieve notification feed",
			slog.Int64("user_id", userID),
			slog.Int("limit", limit),
			slog.Int("page", page),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.log.Info("User notification feed retrieved",
		slog.Int64("user_id", userID),
		slog.Int("count", len(notifications)),
		slog.Int("page", page),
	)

	return notifications, nil
}

func (s *Service) ReadNotification(ctx context.Context, id int64) error {
	if id <= 0 {
		s.log.Error("Invalid notification ID", slog.Int64("id", id))
		return custom_errors.ErrInvalidInput
	}

	s.log.Info("Reading notification", slog.Int64("id", id))

	err := s.notificationRepo.MarkAsRead(ctx, id)
	if err != nil {
		if errors.Is(err, custom_errors.ErrNotificationNotFound) {
			s.log.Debug("Notification not found", slog.Int64("id", id))
			return custom_errors.ErrNotificationNotFound
		}

		s.log.Error("Failed to read notification",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.log.Info("Notification marked as read", slog.Int64("id", id))
	return nil
}

func (s *Service) ReadAllUserNotifications(ctx context.Context, userID int64) error {
	if userID <= 0 {
		s.log.Error("Invalid user ID", slog.Int64("user_id", userID))
		return custom_errors.ErrInvalidInput
	}

	s.log.Info("Reading all notifications for user", slog.Int64("user_id", userID))

	err := s.notificationRepo.MarkAllAsRead(ctx, userID)
	if err != nil {
		s.log.Error("Failed to mark all notifications as read",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.log.Info("All user notifications marked as read", slog.Int64("user_id", userID))
	return nil
}

func (s *Service) RemoveNotification(ctx context.Context, id int64) error {
	if id <= 0 {
		s.log.Error("Invalid notification ID", slog.Int64("id", id))
		return custom_errors.ErrInvalidInput
	}

	s.log.Info("Removing notification", slog.Int64("id", id))

	err := s.notificationRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, custom_errors.ErrNotificationNotFound) {
			s.log.Debug("Notification not found", slog.Int64("id", id))
			return custom_errors.ErrNotificationNotFound
		}

		s.log.Error("Failed to remove notification",
			slog.Int64("id", id),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.log.Info("Notification removed successfully", slog.Int64("id", id))
	return nil
}
