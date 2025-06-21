package notification_grpc

import (
	"context"
	"errors"
	"log/slog"
	"pinstack-notification-service/internal/model"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	notification_service "pinstack-notification-service/internal/service/notification"
)

type NotificationDetailsGetter interface {
	GetNotificationDetails(ctx context.Context, id int64) (*model.Notification, error)
}

type GetNotificationDetailsHandler struct {
	notificationService NotificationDetailsGetter
	log                 *logger.Logger
}

func NewGetNotificationDetailsHandler(
	notificationService notification_service.NotificationService,
	log *logger.Logger,
) *GetNotificationDetailsHandler {
	return &GetNotificationDetailsHandler{
		notificationService: notificationService,
		log:                 log,
	}
}

type NotificationDetailsRequestInternal struct {
	NotificationID int64 `validate:"required,gt=0"`
}

func (h *GetNotificationDetailsHandler) Handle(ctx context.Context, req *pb.GetNotificationDetailsRequest) (*pb.NotificationResponse, error) {
	h.log.Info("Processing get notification details request", slog.Int64("notification_id", req.GetNotificationId()))

	validationReq := &NotificationDetailsRequestInternal{
		NotificationID: req.GetNotificationId(),
	}

	if err := validate.Struct(validationReq); err != nil {
		h.log.Error("Validation failed for get notification details request",
			slog.Int64("notification_id", req.GetNotificationId()),
			slog.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	notification, err := h.notificationService.GetNotificationDetails(ctx, req.GetNotificationId())
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			h.log.Error("Invalid input for get notification details",
				slog.Int64("notification_id", req.GetNotificationId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrNotificationNotFound):
			h.log.Error("Notification not found",
				slog.Int64("notification_id", req.GetNotificationId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, custom_errors.ErrNotificationNotFound.Error())
		default:
			h.log.Error("Internal service error while getting notification details",
				slog.Int64("notification_id", req.GetNotificationId()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	h.log.Info("Successfully retrieved notification details",
		slog.Int64("notification_id", notification.ID),
		slog.Int64("user_id", notification.UserID),
		slog.String("notification_type", string(notification.Type)))

	return &pb.NotificationResponse{
		Id:        notification.ID,
		UserId:    notification.UserID,
		Type:      string(notification.Type),
		IsRead:    notification.IsRead,
		CreatedAt: timestamppb.New(notification.CreatedAt),
		Payload:   notification.Payload,
	}, nil
}
