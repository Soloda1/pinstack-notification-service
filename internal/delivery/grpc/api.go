package notification_grpc

import (
	"github.com/go-playground/validator/v10"
	"pinstack-notification-service/internal/logger"
	notification_service "pinstack-notification-service/internal/service/notification"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
)

var validate = validator.New()

type NotificationGRPCService struct {
	pb.UnimplementedNotificationServiceServer
	notificationService notification_service.NotificationService
	log                 *logger.Logger
}

func NewNotificationGRPCService(notificationService notification_service.NotificationService, log *logger.Logger) *NotificationGRPCService {

	return &NotificationGRPCService{
		notificationService: notificationService,
		log:                 log,
	}
}
