package notification_grpc_test

import (
	"context"
	"encoding/json"
	"errors"
	model "pinstack-notification-service/internal/domain/models"
	notification_grpc "pinstack-notification-service/internal/infrastructure/inbound/grpc"
	"pinstack-notification-service/internal/infrastructure/logger"
	"pinstack-notification-service/mocks"
	"testing"
	"time"

	"github.com/soloda1/pinstack-proto-definitions/custom_errors"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetUserNotificationFeed_Success(t *testing.T) {
	testTime := time.Now()
	payload, _ := json.Marshal(map[string]interface{}{"key": "value"})

	mockService := mocks.NewNotificationService(t)
	log := logger.New("dev")

	notifications := []*model.Notification{
		{
			ID:        1,
			UserID:    1,
			Type:      "test_notification_1",
			IsRead:    false,
			CreatedAt: testTime,
			Payload:   json.RawMessage(payload),
		},
		{
			ID:        2,
			UserID:    1,
			Type:      "test_notification_2",
			IsRead:    true,
			CreatedAt: testTime,
			Payload:   json.RawMessage(payload),
		},
	}

	mockService.On("GetUserNotificationFeed", mock.Anything, int64(1), 10, 1).Return(notifications, int32(2), nil)

	handler := notification_grpc.NewGetUserNotificationFeedHandler(mockService, log)
	req := &pb.GetUserNotificationFeedRequest{
		UserId: 1,
		Limit:  10,
		Page:   1,
	}

	resp, err := handler.Handle(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Notifications))
	assert.Equal(t, int32(2), resp.Total)
	assert.Equal(t, notifications[0].ID, resp.Notifications[0].Id)
	assert.Equal(t, string(notifications[0].Type), resp.Notifications[0].Type)

	mockService.AssertExpectations(t)
}

func TestGetUserNotificationFeed_EmptyList(t *testing.T) {
	mockService := mocks.NewNotificationService(t)
	log := logger.New("dev")

	mockService.On("GetUserNotificationFeed", mock.Anything, int64(1), 10, 1).Return([]*model.Notification{}, int32(0), nil)

	handler := notification_grpc.NewGetUserNotificationFeedHandler(mockService, log)
	req := &pb.GetUserNotificationFeedRequest{
		UserId: 1,
		Limit:  10,
		Page:   1,
	}

	resp, err := handler.Handle(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 0, len(resp.Notifications))
	assert.Equal(t, int32(0), resp.Total)

	mockService.AssertExpectations(t)
}

func TestGetUserNotificationFeed_UserNotFound(t *testing.T) {
	mockService := mocks.NewNotificationService(t)
	log := logger.New("dev")

	mockService.On("GetUserNotificationFeed", mock.Anything, int64(999), 10, 1).Return(nil, int32(0), custom_errors.ErrUserNotFound)

	handler := notification_grpc.NewGetUserNotificationFeedHandler(mockService, log)
	req := &pb.GetUserNotificationFeedRequest{
		UserId: 999,
		Limit:  10,
		Page:   1,
	}

	resp, err := handler.Handle(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)
	statusErr, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, statusErr.Code())
	assert.Contains(t, statusErr.Message(), "user not found")

	mockService.AssertExpectations(t)
}

func TestGetUserNotificationFeed_InternalError(t *testing.T) {
	mockService := mocks.NewNotificationService(t)
	log := logger.New("dev")

	mockService.On("GetUserNotificationFeed", mock.Anything, int64(1), 10, 1).Return(nil, int32(0), errors.New("database error"))

	handler := notification_grpc.NewGetUserNotificationFeedHandler(mockService, log)
	req := &pb.GetUserNotificationFeedRequest{
		UserId: 1,
		Limit:  10,
		Page:   1,
	}

	resp, err := handler.Handle(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, resp)
	statusErr, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, statusErr.Code())
	assert.Contains(t, statusErr.Message(), custom_errors.ErrExternalServiceError.Error())

	mockService.AssertExpectations(t)
}
