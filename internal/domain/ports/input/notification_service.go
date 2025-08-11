package input

import (
	"context"
	"pinstack-notification-service/internal/domain/models"
)

//go:generate mockery --name=NotificationService --output=../../../mocks --outpkg=mocks --case=underscore --with-expecter
type NotificationService interface {
	SaveNotification(ctx context.Context, notification *models.Notification) (int64, error)
	GetNotificationDetails(ctx context.Context, id int64) (*models.Notification, error)
	GetUserNotificationFeed(ctx context.Context, userID int64, limit, page int) ([]*models.Notification, int32, error)
	ReadNotification(ctx context.Context, id int64) error
	ReadAllUserNotifications(ctx context.Context, userID int64) error
	RemoveNotification(ctx context.Context, id int64) error
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
}
