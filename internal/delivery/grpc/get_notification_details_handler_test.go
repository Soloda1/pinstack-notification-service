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
	"time"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGetNotificationDetailsHandler_Handle(t *testing.T) {
	testTime := time.Now()
	payload, _ := json.Marshal(map[string]interface{}{"key": "value"})

	tests := []struct {
		name           string
		req            *pb.GetNotificationDetailsRequest
		mockSetup      func(*mocks.NotificationService)
		wantErr        bool
		expectedCode   codes.Code
		expectedErrMsg string
		expectedResp   *pb.NotificationResponse
	}{
		{
			name: "successful get notification details",
			req: &pb.GetNotificationDetailsRequest{
				NotificationId: 1,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("GetNotificationDetails", mock.Anything, int64(1)).Return(&model.Notification{
					ID:        1,
					UserID:    2,
					Type:      "test_notification",
					IsRead:    false,
					CreatedAt: testTime,
					Payload:   json.RawMessage(payload),
				}, nil)
			},
			wantErr: false,
			expectedResp: &pb.NotificationResponse{
				Id:      1,
				UserId:  2,
				Type:    "test_notification",
				IsRead:  false,
				Payload: payload,
			},
		},
		{
			name: "validation error - notification ID zero",
			req: &pb.GetNotificationDetailsRequest{
				NotificationId: 0,
			},
			mockSetup:      func(mockService *mocks.NotificationService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "validation failed",
		},
		{
			name: "notification not found",
			req: &pb.GetNotificationDetailsRequest{
				NotificationId: 999,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("GetNotificationDetails", mock.Anything, int64(999)).Return(nil, custom_errors.ErrNotificationNotFound)
			},
			wantErr:        true,
			expectedCode:   codes.NotFound,
			expectedErrMsg: "notification not found",
		},
		{
			name: "invalid input",
			req: &pb.GetNotificationDetailsRequest{
				NotificationId: -1,
			},
			mockSetup:      func(mockService *mocks.NotificationService) {},
			wantErr:        true,
			expectedCode:   codes.InvalidArgument,
			expectedErrMsg: "validation failed",
		},
		{
			name: "internal service error",
			req: &pb.GetNotificationDetailsRequest{
				NotificationId: 1,
			},
			mockSetup: func(mockService *mocks.NotificationService) {
				mockService.On("GetNotificationDetails", mock.Anything, int64(1)).Return(nil, errors.New("database error"))
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

			handler := notification_grpc.NewGetNotificationDetailsHandler(mockService, log)
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

				assert.Equal(t, tt.expectedResp.Id, resp.Id)
				assert.Equal(t, tt.expectedResp.UserId, resp.UserId)
				assert.Equal(t, tt.expectedResp.Type, resp.Type)
				assert.Equal(t, tt.expectedResp.IsRead, resp.IsRead)
				assert.NotNil(t, resp.CreatedAt)
				assert.Equal(t, string(tt.expectedResp.Payload), string(resp.Payload))
			}

			mockService.AssertExpectations(t)
		})
	}
}
