package notification_service_test

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	"pinstack-notification-service/internal/model"
	notification_service "pinstack-notification-service/internal/service/notification"
	"pinstack-notification-service/mocks"
	"testing"
	"time"
)

func TestService_SendNotification(t *testing.T) {
	tests := []struct {
		name         string
		notification *model.Notification
		mockSetup    func(*mocks.NotificationRepository)
		wantErr      bool
		expectedErr  error
	}{
		{
			name: "successful notification send",
			notification: &model.Notification{
				UserID:  1,
				Type:    "relation",
				IsRead:  false,
				Payload: json.RawMessage(`{"event":"new_follower","follower_id":42}`),
			},
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(notif *model.Notification) bool {
					return notif.UserID == 1 && notif.Type == "relation" && !notif.IsRead
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			notification: &model.Notification{
				UserID:  1,
				Type:    "relation",
				IsRead:  false,
				Payload: json.RawMessage(`{"event":"new_follower","follower_id":42}`),
			},
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("Create", mock.Anything, mock.Anything).Return(custom_errors.ErrDatabaseQuery)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:         "nil notification",
			notification: nil,
			mockSetup:    func(repo *mocks.NotificationRepository) {},
			wantErr:      true,
			expectedErr:  custom_errors.ErrInvalidInput,
		},
		{
			name: "invalid user ID",
			notification: &model.Notification{
				UserID:  0, // Invalid user ID
				Type:    "relation",
				IsRead:  false,
				Payload: json.RawMessage(`{"event":"new_follower","follower_id":42}`),
			},
			mockSetup:   func(repo *mocks.NotificationRepository) {},
			wantErr:     true,
			expectedErr: custom_errors.ErrInvalidInput,
		},
		{
			name: "empty notification type",
			notification: &model.Notification{
				UserID:  1,
				Type:    "", // Empty type
				IsRead:  false,
				Payload: json.RawMessage(`{"event":"new_follower","follower_id":42}`),
			},
			mockSetup:   func(repo *mocks.NotificationRepository) {},
			wantErr:     true,
			expectedErr: custom_errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewNotificationRepository(t)
			log := logger.New("dev")

			tt.mockSetup(mockRepo)

			service := notification_service.NewNotificationService(log, mockRepo)
			err := service.SendNotification(context.Background(), tt.notification)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetNotificationDetails(t *testing.T) {
	createdAt := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"event":"new_follower","follower_id":42}`)

	notif := &model.Notification{
		ID:        1,
		UserID:    2,
		Type:      "relation",
		IsRead:    false,
		CreatedAt: createdAt,
		Payload:   payload,
	}

	tests := []struct {
		name        string
		id          int64
		mockSetup   func(*mocks.NotificationRepository)
		want        *model.Notification
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful get notification",
			id:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("GetByID", mock.Anything, int64(1)).Return(notif, nil)
			},
			want:    notif,
			wantErr: false,
		},
		{
			name: "notification not found",
			id:   999,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("GetByID", mock.Anything, int64(999)).Return(nil, custom_errors.ErrNotificationNotFound)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrNotificationNotFound,
		},
		{
			name: "repository error",
			id:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("GetByID", mock.Anything, int64(1)).Return(nil, custom_errors.ErrDatabaseQuery)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:        "invalid id",
			id:          0, // Invalid ID
			mockSetup:   func(repo *mocks.NotificationRepository) {},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewNotificationRepository(t)
			log := logger.New("dev")

			tt.mockSetup(mockRepo)

			service := notification_service.NewNotificationService(log, mockRepo)
			got, err := service.GetNotificationDetails(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_GetUnreadCount(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		mockSetup   func(*mocks.NotificationRepository)
		want        int
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successful get unread count",
			userID: 1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("CountUnread", mock.Anything, int64(1)).Return(5, nil)
			},
			want:    5,
			wantErr: false,
		},
		{
			name:   "zero unread count",
			userID: 2,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("CountUnread", mock.Anything, int64(2)).Return(0, nil)
			},
			want:    0,
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("CountUnread", mock.Anything, int64(1)).Return(0, custom_errors.ErrDatabaseQuery)
			},
			want:        0,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:        "invalid user ID",
			userID:      0, // Invalid user ID
			mockSetup:   func(repo *mocks.NotificationRepository) {},
			want:        0,
			wantErr:     true,
			expectedErr: custom_errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewNotificationRepository(t)
			log := logger.New("dev")

			tt.mockSetup(mockRepo)

			service := notification_service.NewNotificationService(log, mockRepo)
			got, err := service.GetUnreadCount(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_GetUserNotificationFeed(t *testing.T) {
	createdAt1 := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	createdAt2 := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	payload1 := json.RawMessage(`{"event":"new_follower","follower_id":42}`)
	payload2 := json.RawMessage(`{"event":"new_follower","follower_id":43}`)

	notifications := []*model.Notification{
		{
			ID:        1,
			UserID:    5,
			Type:      "relation",
			IsRead:    false,
			CreatedAt: createdAt1,
			Payload:   payload1,
		},
		{
			ID:        2,
			UserID:    5,
			Type:      "relation",
			IsRead:    true,
			CreatedAt: createdAt2,
			Payload:   payload2,
		},
	}

	tests := []struct {
		name        string
		userID      int64
		limit       int
		page        int
		mockSetup   func(*mocks.NotificationRepository)
		want        []*model.Notification
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successful get notifications",
			userID: 5,
			limit:  10,
			page:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				// Page 1, limit 10 should result in offset 0
				repo.On("ListByUser", mock.Anything, int64(5), 10, 0).Return(notifications, nil)
			},
			want:    notifications,
			wantErr: false,
		},
		{
			name:   "get second page",
			userID: 5,
			limit:  5,
			page:   2,
			mockSetup: func(repo *mocks.NotificationRepository) {
				// Page 2, limit 5 should result in offset 5
				repo.On("ListByUser", mock.Anything, int64(5), 5, 5).Return(notifications[1:], nil)
			},
			want:    notifications[1:],
			wantErr: false,
		},
		{
			name:   "handle negative limit",
			userID: 5,
			limit:  -1,
			page:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				// Negative limit should be changed to default (10)
				repo.On("ListByUser", mock.Anything, int64(5), 10, 0).Return(notifications, nil)
			},
			want:    notifications,
			wantErr: false,
		},
		{
			name:   "handle negative page",
			userID: 5,
			limit:  10,
			page:   -1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				// Negative page should be changed to 1
				repo.On("ListByUser", mock.Anything, int64(5), 10, 0).Return(notifications, nil)
			},
			want:    notifications,
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 5,
			limit:  10,
			page:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("ListByUser", mock.Anything, int64(5), 10, 0).Return(nil, custom_errors.ErrDatabaseQuery)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:        "invalid user ID",
			userID:      0, // Invalid user ID
			limit:       10,
			page:        1,
			mockSetup:   func(repo *mocks.NotificationRepository) {},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewNotificationRepository(t)
			log := logger.New("dev")

			tt.mockSetup(mockRepo)

			service := notification_service.NewNotificationService(log, mockRepo)
			got, err := service.GetUserNotificationFeed(context.Background(), tt.userID, tt.limit, tt.page)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_ReadNotification(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		mockSetup   func(*mocks.NotificationRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful mark as read",
			id:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("MarkAsRead", mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "notification not found",
			id:   999,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("MarkAsRead", mock.Anything, int64(999)).Return(custom_errors.ErrNotificationNotFound)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrNotificationNotFound,
		},
		{
			name: "repository error",
			id:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("MarkAsRead", mock.Anything, int64(1)).Return(custom_errors.ErrDatabaseQuery)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:        "invalid id",
			id:          0, // Invalid ID
			mockSetup:   func(repo *mocks.NotificationRepository) {},
			wantErr:     true,
			expectedErr: custom_errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewNotificationRepository(t)
			log := logger.New("dev")

			tt.mockSetup(mockRepo)

			service := notification_service.NewNotificationService(log, mockRepo)
			err := service.ReadNotification(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ReadAllUserNotifications(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		mockSetup   func(*mocks.NotificationRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successful mark all as read",
			userID: 1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("MarkAllAsRead", mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("MarkAllAsRead", mock.Anything, int64(1)).Return(custom_errors.ErrDatabaseQuery)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:        "invalid user ID",
			userID:      0, // Invalid user ID
			mockSetup:   func(repo *mocks.NotificationRepository) {},
			wantErr:     true,
			expectedErr: custom_errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewNotificationRepository(t)
			log := logger.New("dev")

			tt.mockSetup(mockRepo)

			service := notification_service.NewNotificationService(log, mockRepo)
			err := service.ReadAllUserNotifications(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_RemoveNotification(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		mockSetup   func(*mocks.NotificationRepository)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful delete",
			id:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("Delete", mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "notification not found",
			id:   999,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("Delete", mock.Anything, int64(999)).Return(custom_errors.ErrNotificationNotFound)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrNotificationNotFound,
		},
		{
			name: "repository error",
			id:   1,
			mockSetup: func(repo *mocks.NotificationRepository) {
				repo.On("Delete", mock.Anything, int64(1)).Return(custom_errors.ErrDatabaseQuery)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
		{
			name:        "invalid id",
			id:          0, // Invalid ID
			mockSetup:   func(repo *mocks.NotificationRepository) {},
			wantErr:     true,
			expectedErr: custom_errors.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewNotificationRepository(t)
			log := logger.New("dev")

			tt.mockSetup(mockRepo)

			service := notification_service.NewNotificationService(log, mockRepo)
			err := service.RemoveNotification(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
