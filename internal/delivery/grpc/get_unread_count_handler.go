package notification_grpc

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	notification_service "pinstack-notification-service/internal/service/notification"
)

type UnreadCountGetter interface {
	GetUnreadCount(ctx context.Context, userID int64) (int, error)
}

type GetUnreadCountHandler struct {
	notificationService UnreadCountGetter
	log                 *logger.Logger
}

func NewGetUnreadCountHandler(
	notificationService notification_service.NotificationService,
	log *logger.Logger,
) *GetUnreadCountHandler {
	return &GetUnreadCountHandler{
		notificationService: notificationService,
		log:                 log,
	}
}

type GetUnreadCountRequestInternal struct {
	UserID int64 `validate:"required,gt=0"`
}

func (h *GetUnreadCountHandler) Handle(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	h.log.Info("Processing get unread count request", slog.Int64("user_id", req.GetUserId()))

	validationReq := &GetUnreadCountRequestInternal{
		UserID: req.GetUserId(),
	}

	if err := validate.Struct(validationReq); err != nil {
		h.log.Error("Validation failed for get unread count request",
			slog.Int64("user_id", req.GetUserId()),
			slog.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	count, err := h.notificationService.GetUnreadCount(ctx, req.GetUserId())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			h.log.Error("Invalid input for get unread count",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrUserNotFound):
			h.log.Error("User not found",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		default:
			h.log.Error("Internal service error while getting unread count",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	h.log.Info("Successfully retrieved unread count",
		slog.Int64("user_id", req.GetUserId()),
		slog.Int("count", count))

	return &pb.GetUnreadCountResponse{
		Count: int32(count),
	}, nil
}
