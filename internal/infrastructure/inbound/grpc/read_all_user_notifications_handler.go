package notification_grpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/soloda1/pinstack-proto-definitions/custom_errors"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	notification_service "pinstack-notification-service/internal/domain/ports/input"
	ports "pinstack-notification-service/internal/domain/ports/output"
)

type AllUserNotificationsReader interface {
	ReadAllUserNotifications(ctx context.Context, userID int64) error
}

type ReadAllUserNotificationsHandler struct {
	notificationService AllUserNotificationsReader
	log                 ports.Logger
}

func NewReadAllUserNotificationsHandler(
	notificationService notification_service.NotificationService,
	log ports.Logger,
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
	h.log.Info("Processing read all user notifications request", slog.Int64("user_id", req.GetUserId()))

	validationReq := &ReadAllUserNotificationsRequestInternal{
		UserID: req.GetUserId(),
	}

	if err := validate.Struct(validationReq); err != nil {
		h.log.Error("Validation failed for read all user notifications request",
			slog.Int64("user_id", req.GetUserId()),
			slog.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	err := h.notificationService.ReadAllUserNotifications(ctx, req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			h.log.Error("Invalid input for read all user notifications",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrUserNotFound):
			h.log.Error("User not found for read all notifications request",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		default:
			h.log.Error("Internal service error while reading all user notifications",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, custom_errors.ErrExternalServiceError.Error())
		}
	}

	h.log.Info("Successfully marked all notifications as read", slog.Int64("user_id", req.GetUserId()))
	return &emptypb.Empty{}, nil
}
