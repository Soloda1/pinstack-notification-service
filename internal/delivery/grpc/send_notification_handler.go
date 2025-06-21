package notification_grpc

import (
	"context"
	"errors"
	"github.com/soloda1/pinstack-proto-definitions/events"
	"pinstack-notification-service/internal/model"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	notification_service "pinstack-notification-service/internal/service/notification"
)

type NotificationSender interface {
	SaveNotification(ctx context.Context, notification *model.Notification) error
}

type SendNotificationHandler struct {
	notificationService NotificationSender
	log                 *logger.Logger
}

func NewSendNotificationHandler(
	notificationService notification_service.NotificationService,
	log *logger.Logger,
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
		"user_id", req.GetUserId(),
		"type", req.GetType(),
		"payload_size", len(req.GetPayload()))

	validationReq := &SendNotificationRequestInternal{
		UserID:  req.GetUserId(),
		Type:    req.GetType(),
		Payload: req.GetPayload(),
	}

	if err := validate.Struct(validationReq); err != nil {
		h.log.Error("Validation failed for send notification request",
			"user_id", req.GetUserId(),
			"type", req.GetType(),
			"error", err)
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	notification := &model.Notification{
		UserID:  req.GetUserId(),
		Type:    events.EventType(req.GetType()),
		IsRead:  false,
		Payload: req.GetPayload(),
	}

	err := h.notificationService.SaveNotification(ctx, notification)
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			h.log.Error("Invalid input for send notification",
				"user_id", req.GetUserId(),
				"type", req.GetType(),
				"error", err)
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrUserNotFound):
			h.log.Error("User not found when sending notification",
				"user_id", req.GetUserId(),
				"type", req.GetType(),
				"error", err)
			return nil, status.Error(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		default:
			h.log.Error("Internal service error while sending notification",
				"user_id", req.GetUserId(),
				"type", req.GetType(),
				"error", err)
			return nil, status.Error(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	h.log.Info("Successfully sent notification",
		"notification_id", notification.ID,
		"user_id", notification.UserID,
		"type", notification.Type)

	return &pb.SendNotificationResponse{
		NotificationId: notification.ID,
	}, nil
}
