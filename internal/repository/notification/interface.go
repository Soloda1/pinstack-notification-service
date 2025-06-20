package notification_repository

import (
	"context"
	"pinstack-notification-service/internal/model"
)

type NotificationRepository interface {
	Create(ctx context.Context, notif *model.Notification) error
	GetByID(ctx context.Context, id int64) (*model.Notification, error)
	ListByUser(ctx context.Context, userID int64, limit int, offset int) ([]*model.Notification, error)
	MarkAsRead(ctx context.Context, id int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	Delete(ctx context.Context, id int64) error
}
