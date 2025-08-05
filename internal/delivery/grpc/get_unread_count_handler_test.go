package notification_grpc_test

import (
	"context"
	"errors"
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	notification_grpc "pinstack-notification-service/internal/delivery/grpc"
	"pinstack-notification-service/internal/logger"
	"pinstack-notification-service/mocks"
	"testing"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetUnreadCountHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		req            *pb.GetUnreadCountRequest
		mockSetup      func(*mocks.NotificationService)
		wantErr        bool
		expectedCode   codes.Code
		expectedErrMsg string
		expectedCount  int32
	}{
		{
			name: "successful get unread count",
			req: &pb.GetUnreadCountRequest{
				UserId: 1,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("GetUnreadCount", mock.Anything, int64(1)).Return(5, nil)
			},
			wantErr:       false,
			expectedCount: 5,
		},
		{
			name: "validation error - user ID zero",
			req: &pb.GetUnreadCountRequest{
				UserId: 0,
			},
			mockSetup:      func(mockService *mocks.NotificationService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "validation failed",
		},
		{
			name: "user not found",
			req: &pb.GetUnreadCountRequest{
				UserId: 999,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("GetUnreadCount", mock.Anything, int64(999)).Return(0, custom_errors.ErrUserNotFound)
			},
			wantErr:        true,
			expectedCode:   codes.NotFound,
			expectedErrMsg: "user not found",
		},
		{
			name: "invalid input",
			req: &pb.GetUnreadCountRequest{
				UserId: -1,
			},
			mockSetup:      func(mockService *mocks.NotificationService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "validation failed",
		},
		{
			name: "internal service error",
			req: &pb.GetUnreadCountRequest{
				UserId: 1,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("GetUnreadCount", mock.Anything, int64(1)).Return(0, errors.New("database error"))
			},
			wantErr:        true,
			expectedCode:   codes.Internal,
			expectedErrMsg: custom_errors.ErrExternalServiceError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewNotificationService(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			handler := notification_grpc.NewGetUnreadCountHandler(mockService, log)
			resp, err := handler.Handle(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				statusErr, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, statusErr.Code())
				assert.Contains(t, statusErr.Message(), tt.expectedErrMsg)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.expectedCount, resp.Count)
			}

			mockService.AssertExpectations(t)
		})
	}
}
