package notification_grpc

import (
	"context"
	"errors"

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
	validationReq := &GetUnreadCountRequestInternal{
		UserID: req.GetUserId(),
	}

	if err := validate.Struct(validationReq); err != nil {
		return nil, status.Error(codes.InvalidArgument, custom_errors.ErrValidationFailed.Error())
	}

	count, err := h.notificationService.GetUnreadCount(ctx, req.GetUserId())
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

	return &pb.GetUnreadCountResponse{
		Count: int32(count),
	}, nil
}
