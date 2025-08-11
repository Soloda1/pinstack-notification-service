package output

import (
	"context"
	"pinstack-notification-service/internal/domain/models"
)

//go:generate mockery --name=NotificationRepository --output=../../../mocks --outpkg=mocks --case=underscore --with-expecter
type NotificationRepository interface {
	Create(ctx context.Context, notif *models.Notification) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Notification, error)
	ListByUser(ctx context.Context, userID int64, limit int, offset int) ([]*models.Notification, int32, error)
	MarkAsRead(ctx context.Context, id int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	Delete(ctx context.Context, id int64) error
	CountUnread(ctx context.Context, userID int64) (int, error)
}
