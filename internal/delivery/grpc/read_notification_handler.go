package notification_grpc

import (
	"context"
	"errors"
	"log/slog"

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
	h.log.Info("Processing read notification request", slog.Int64("notification_id", req.GetNotificationId()))

	validationReq := &ReadNotificationRequestInternal{
		NotificationID: req.GetNotificationId(),
	}

	if err := validate.Struct(validationReq); err != nil {
		h.log.Error("Validation failed for read notification request",
			slog.Int64("notification_id", req.GetNotificationId()),
			slog.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	err := h.notificationService.ReadNotification(ctx, req.GetNotificationId())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			h.log.Error("Invalid input for read notification",
				slog.Int64("notification_id", req.GetNotificationId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrNotificationNotFound):
			h.log.Error("Notification not found",
				slog.Int64("notification_id", req.GetNotificationId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, custom_errors.ErrNotificationNotFound.Error())
		default:
			h.log.Error("Internal service error while reading notification",
				slog.Int64("notification_id", req.GetNotificationId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	h.log.Info("Successfully marked notification as read", slog.Int64("notification_id", req.GetNotificationId()))
	return &emptypb.Empty{}, nil
}
