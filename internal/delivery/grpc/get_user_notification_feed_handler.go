package notification_grpc

import (
	"context"
	"errors"
	"pinstack-notification-service/internal/model"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	notification_service "pinstack-notification-service/internal/service/notification"
)

type UserNotificationFeedGetter interface {
	GetUserNotificationFeed(ctx context.Context, userID int64, limit, page int) ([]*model.Notification, error)
}

type GetUserNotificationFeedHandler struct {
	notificationService UserNotificationFeedGetter
	log                 *logger.Logger
}

func NewGetUserNotificationFeedHandler(
	notificationService notification_service.NotificationService,
	log *logger.Logger,
) *GetUserNotificationFeedHandler {
	return &GetUserNotificationFeedHandler{
		notificationService: notificationService,
		log:                 log,
	}
}

type UserNotificationFeedRequestInternal struct {
	UserID int64 `validate:"required,gt=0"`
	Limit  int   `validate:"required,gt=0,lte=100"`
	Page   int   `validate:"required,gte=0"`
}

func (h *GetUserNotificationFeedHandler) Handle(ctx context.Context, req *pb.GetUserNotificationFeedRequest) (*pb.GetUserNotificationFeedResponse, error) {
	h.log.Info("Processing get user notification feed request",
		"user_id", req.GetUserId(),
		"limit", req.GetLimit(),
		"page", req.GetPage())

	validationReq := &UserNotificationFeedRequestInternal{
		UserID: req.GetUserId(),
		Limit:  int(req.GetLimit()),
		Page:   int(req.GetPage()),
	}

	if err := validate.Struct(validationReq); err != nil {
		h.log.Error("Validation failed for user notification feed request",
			"user_id", req.GetUserId(),
			"limit", req.GetLimit(),
			"page", req.GetPage(),
			"error", err)
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	notifications, err := h.notificationService.GetUserNotificationFeed(ctx, req.GetUserId(), int(req.GetLimit()), int(req.GetPage()))
	if err != nil {
		switch {
		case errors.Is(err, custom_errors.ErrInvalidInput):
			h.log.Error("Invalid input for get user notification feed",
				"user_id", req.GetUserId(),
				"limit", req.GetLimit(),
				"page", req.GetPage(),
				"error", err)
			return nil, status.Error(codes.InvalidArgument, custom_errors.ErrInvalidInput.Error())
		case errors.Is(err, custom_errors.ErrUserNotFound):
			h.log.Error("User not found for notification feed request",
				"user_id", req.GetUserId(),
				"error", err)
			return nil, status.Error(codes.NotFound, custom_errors.ErrUserNotFound.Error())
		default:
			h.log.Error("Internal service error while getting user notification feed",
				"user_id", req.GetUserId(),
				"limit", req.GetLimit(),
				"page", req.GetPage(),
				"error", err)
			return nil, status.Error(codes.Internal, custom_errors.ErrInternalServiceError.Error())
		}
	}

	response := &pb.GetUserNotificationFeedResponse{
		Notifications: make([]*pb.NotificationResponse, 0, len(notifications)),
	}

	for _, notification := range notifications {
		response.Notifications = append(response.Notifications, &pb.NotificationResponse{
			Id:        notification.ID,
			UserId:    notification.UserID,
			Type:      string(notification.Type),
			IsRead:    notification.IsRead,
			CreatedAt: timestamppb.New(notification.CreatedAt),
			Payload:   notification.Payload,
		})
	}

	h.log.Info("Successfully retrieved user notification feed",
		"user_id", req.GetUserId(),
		"notifications_count", len(notifications))

	return response, nil
}
