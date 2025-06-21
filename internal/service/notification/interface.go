package notification_service

import (
	"context"
	"pinstack-notification-service/internal/model"
)

//go:generate mockery --name=NotificationService --output=../../../mocks --outpkg=mocks --case=underscore --with-expecter
type NotificationService interface {
	SaveNotification(ctx context.Context, notification *model.Notification) error
	GetNotificationDetails(ctx context.Context, id int64) (*model.Notification, error)
	GetUserNotificationFeed(ctx context.Context, userID int64, limit, page int) ([]*model.Notification, error)
	ReadNotification(ctx context.Context, id int64) error
	ReadAllUserNotifications(ctx context.Context, userID int64) error
	RemoveNotification(ctx context.Context, id int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
}
