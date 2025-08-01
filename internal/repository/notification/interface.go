package notification_repository

import (
	"context"
	"pinstack-notification-service/internal/model"
)

//go:generate mockery --name=NotificationRepository --output=../../../mocks --outpkg=mocks --case=underscore --with-expecter
type NotificationRepository interface {
	Create(ctx context.Context, notif *model.Notification) (int64, error)
	GetByID(ctx context.Context, id int64) (*model.Notification, error)
	ListByUser(ctx context.Context, userID int64, limit int, offset int) ([]*model.Notification, error)
	MarkAsRead(ctx context.Context, id int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	Delete(ctx context.Context, id int64) error
	CountUnread(ctx context.Context, userID int64) (int, error)
}
