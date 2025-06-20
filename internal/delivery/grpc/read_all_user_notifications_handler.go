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

type AllUserNotificationsReader interface {
	ReadAllUserNotifications(ctx context.Context, userID int64) error
}

type ReadAllUserNotificationsHandler struct {
	notificationService AllUserNotificationsReader
	log                 *logger.Logger
}

func NewReadAllUserNotificationsHandler(
	notificationService notification_service.NotificationService,
	log *logger.Logger,
) *ReadAllUserNotificationsHandler {
	return &ReadAllUserNotificationsHandler{
		notificationService: notificationService,
		log:                 log,
	}
}

type ReadAllUserNotificationsRequestInternal struct {
	UserID int64 `validate:"required,gt=0"`
}

func (h *ReadAllUserNotificationsHandler) Handle(ctx context.Context, req *pb.ReadAllUserNotificationsRequest) (*emptypb.Empty, error) {
	validationReq := &ReadAllUserNotificationsRequestInternal{
		UserID: req.GetUserId(),
	}

	if err := validate.Struct(validationReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	err := h.notificationService.ReadAllUserNotifications(ctx, req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrUserNotFound):
			return nil, status.Error(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		default:
			return nil, status.Error(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	return &emptypb.Empty{}, nil
}
