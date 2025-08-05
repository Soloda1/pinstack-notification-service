package notification_grpc_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	notification_grpc "pinstack-notification-service/internal/delivery/grpc"
	"pinstack-notification-service/internal/logger"
	"pinstack-notification-service/internal/model"
	"pinstack-notification-service/mocks"
	"testing"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSendNotificationHandler_Handle(t *testing.T) {
	payload, _ := json.Marshal(map[string]interface{}{"key": "value"})

	tests := []struct {
		name           string
		req            *pb.SendNotificationRequest
		mockSetup      func(*mocks.NotificationService)
		wantErr        bool
		expectedCode   codes.Code
		expectedErrMsg string
		expectedID     int64
	}{
		{
			name: "successful send notification",
			req: &pb.SendNotificationRequest{
				UserId:  1,
				Type:    "test_notification",
				Payload: payload,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("SaveNotification", mock.Anything, mock.MatchedBy(func(n *model.Notification) bool {
					return n.UserID == 1 && n.Type == "test_notification" && string(n.Payload) == string(payload)
				})).Run(func(args mock.Arguments) {
					notification := args.Get(1).(*model.Notification)
					notification.ID = 100
				}).Return(int64(100), nil)
			},
			wantErr:    false,
			expectedID: 100,
		},
		{
			name: "validation error - user ID zero",
			req: &pb.SendNotificationRequest{
				UserId:  0,
				Type:    "test_notification",
				Payload: payload,
			},
			mockSetup:      func(mockService *mocks.NotificationService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "validation failed",
		},
		{
			name: "validation error - empty type",
			req: &pb.SendNotificationRequest{
				UserId:  1,
				Type:    "",
				Payload: payload,
			},
			mockSetup:      func(mockService *mocks.NotificationService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "validation failed",
		},
		{
			name: "validation error - nil payload",
			req: &pb.SendNotificationRequest{
				UserId:  1,
				Type:    "test_notification",
				Payload: nil,
			},
			mockSetup:      func(mockService *mocks.NotificationService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "validation failed",
		},
		{
			name: "user not found",
			req: &pb.SendNotificationRequest{
				UserId:  999,
				Type:    "test_notification",
				Payload: payload,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("SaveNotification", mock.Anything, mock.MatchedBy(func(n *model.Notification) bool {
					return n.UserID == 999 && n.Type == "test_notification" && string(n.Payload) == string(payload)
				})).Return(int64(0), custom_errors.ErrUserNotFound)
			},
			wantErr:        true,
			expectedCode:   codes.NotFound,
			expectedErrMsg: "user not found",
		},
		{
			name: "invalid input - negative user ID",
			req: &pb.SendNotificationRequest{
				UserId:  -1,
				Type:    "test_notification",
				Payload: payload,
			},
			mockSetup:      func(mockService *mocks.NotificationService) {}, // не ожидаем вызова SaveNotification
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "validation failed",
		},
		{
			name: "internal service error",
			req: &pb.SendNotificationRequest{
				UserId:  1,
				Type:    "test_notification",
				Payload: payload,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("SaveNotification", mock.Anything, mock.Anything).Return(int64(0), errors.New("database error"))
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

			handler := notification_grpc.NewSendNotificationHandler(mockService, log)
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
				assert.Equal(t, tt.expectedID, resp.NotificationId)
			}

			mockService.AssertExpectations(t)
		})
	}
}
