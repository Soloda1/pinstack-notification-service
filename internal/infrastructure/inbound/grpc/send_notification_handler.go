package notification_grpc

import (
	"context"
	"errors"
	"log/slog"
	model "pinstack-notification-service/internal/domain/models"

	"github.com/soloda1/pinstack-proto-definitions/custom_errors"

	"github.com/soloda1/pinstack-proto-definitions/events"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	notification_service "pinstack-notification-service/internal/domain/ports/input"
	ports "pinstack-notification-service/internal/domain/ports/output"
)

type NotificationSender interface {
	SaveNotification(ctx context.Context, notification *model.Notification) (int64, error)
}

type SendNotificationHandler struct {
	notificationService NotificationSender
	log                 ports.Logger
}

func NewSendNotificationHandler(
	notificationService notification_service.NotificationService,
	log ports.Logger,
) *SendNotificationHandler {
	return &SendNotificationHandler{
		notificationService: notificationService,
		log:                 log,
	}
}

type SendNotificationRequestInternal struct {
	UserID  int64  `validate:"required,gt=0"`
	Type    string `validate:"required"`
	Payload []byte `validate:"required"`
}

func (h *SendNotificationHandler) Handle(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	h.log.Info("Processing send notification request",
		slog.Int64("user_id", req.GetUserId()),
		slog.String("type", req.GetType()),
		slog.Int("payload_size", len(req.GetPayload())))

	validationReq := &SendNotificationRequestInternal{
		UserID:  req.GetUserId(),
		Type:    req.GetType(),
		Payload: req.GetPayload(),
	}

	if err := validate.Struct(validationReq); err != nil {
		h.log.Error("Validation failed for send notification request",
			slog.Int64("user_id", req.GetUserId()),
			slog.String("type", req.GetType()),
			slog.String("error", err.Error()))
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	notification := &model.Notification{
		UserID:  req.GetUserId(),
		Type:    events.EventType(req.GetType()),
		IsRead:  false,
		Payload: req.GetPayload(),
	}

	notificationID, err := h.notificationService.SaveNotification(ctx, notification)
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			h.log.Error("Invalid input for send notification",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("type", req.GetType()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrUserNotFound):
			h.log.Error("User not found when sending notification",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("type", req.GetType()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		default:
			h.log.Error("Internal service error while sending notification",
				slog.Int64("user_id", req.GetUserId()),
				slog.String("type", req.GetType()),
				slog.String("error", err.Error()))
			return nil, status.Error(codes.Internal, custom_errors.ErrExternalServiceError.Error())
		}
	}

	h.log.Info("Successfully sent notification",
		slog.Int64("notification_id", notificationID),
		slog.Int64("user_id", notification.UserID),
		slog.String("type", string(notification.Type)))

	return &pb.SendNotificationResponse{
		NotificationId: notificationID,
	}, nil
}
