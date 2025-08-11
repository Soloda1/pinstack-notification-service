package output

import (
	"context"
	"pinstack-notification-service/internal/domain/models"
)

//go:generate mockery --name Client --dir . --output ../../../mocks --outpkg mocks --with-expecter --filename UserClient.go
type Client interface {
	GetUser(ctx context.Context, id int64) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}
