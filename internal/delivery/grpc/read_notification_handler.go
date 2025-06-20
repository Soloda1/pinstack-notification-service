package notification_grpc

import (
	"context"
	"errors"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	notification_service "pinstack-notification-service/internal/service/notification"
)

type NotificationReader interface {
	ReadNotification(ctx context.Context, id int64) error
}

type ReadNotificationHandler struct {
	notificationService NotificationReader
	log                 *logger.Logger
}

func NewReadNotificationHandler(
	notificationService notification_service.NotificationService,
	log *logger.Logger,
) *ReadNotificationHandler {
	return &ReadNotificationHandler{
		notificationService: notificationService,
		log:                 log,
	}
}

type ReadNotificationRequestInternal struct {
	NotificationID int64 `validate:"required,gt=0"`
}

func (h *ReadNotificationHandler) Handle(ctx context.Context, req *pb.ReadNotificationRequest) (*emptypb.Empty, error) {
	validationReq := &ReadNotificationRequestInternal{
		NotificationID: req.GetNotificationId(),
	}

	if err := validate.Struct(validationReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	err := h.notificationService.ReadNotification(ctx, req.GetNotificationId())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrNotificationNotFound):
			return nil, status.Error(codes.NotFound, custom_errors.ErrNotificationNotFound.Error())
		default:
			return nil, status.Error(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	return &emptypb.Empty{}, nil
}
