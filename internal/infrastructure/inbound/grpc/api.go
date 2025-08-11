package notification_grpc

import (
	"context"
	notification_service "pinstack-notification-service/internal/domain/ports/input"
	ports "pinstack-notification-service/internal/domain/ports/output"

	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/soloda1/pinstack-proto-definitions/gen/go/pinstack-proto-definitions/notification/v1"
)

var validate = validator.New()

type NotificationGRPCService struct {
	pb.UnimplementedNotificationServiceServer
	notificationService             notification_service.NotificationService
	log                             ports.Logger
	sendNotificationHandler         *SendNotificationHandler
	getNotificationDetailsHandler   *GetNotificationDetailsHandler
	getUserNotificationFeedHandler  *GetUserNotificationFeedHandler
	readNotificationHandler         *ReadNotificationHandler
	readAllUserNotificationsHandler *ReadAllUserNotificationsHandler
	removeNotificationHandler       *RemoveNotificationHandler
	getUnreadCountHandler           *GetUnreadCountHandler
}

func NewNotificationGRPCService(notificationService notification_service.NotificationService, log ports.Logger) *NotificationGRPCService {
	service := &NotificationGRPCService{
		notificationService: notificationService,
		log:                 log,
	}

	service.sendNotificationHandler = NewSendNotificationHandler(notificationService, log)
	service.getNotificationDetailsHandler = NewGetNotificationDetailsHandler(notificationService, log)
	service.getUserNotificationFeedHandler = NewGetUserNotificationFeedHandler(notificationService, log)
	service.readNotificationHandler = NewReadNotificationHandler(notificationService, log)
	service.readAllUserNotificationsHandler = NewReadAllUserNotificationsHandler(notificationService, log)
	service.removeNotificationHandler = NewRemoveNotificationHandler(notificationService, log)
	service.getUnreadCountHandler = NewGetUnreadCountHandler(notificationService, log)

	return service
}

func (s *NotificationGRPCService) SendNotification(ctx context.Context, req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	return s.sendNotificationHandler.Handle(ctx, req)
}

func (s *NotificationGRPCService) GetNotificationDetails(ctx context.Context, req *pb.GetNotificationDetailsRequest) (*pb.NotificationResponse, error) {
	return s.getNotificationDetailsHandler.Handle(ctx, req)
}

func (s *NotificationGRPCService) GetUserNotificationFeed(ctx context.Context, req *pb.GetUserNotificationFeedRequest) (*pb.GetUserNotificationFeedResponse, error) {
	return s.getUserNotificationFeedHandler.Handle(ctx, req)
}

func (s *NotificationGRPCService) ReadNotification(ctx context.Context, req *pb.ReadNotificationRequest) (*emptypb.Empty, error) {
	return s.readNotificationHandler.Handle(ctx, req)
}

func (s *NotificationGRPCService) ReadAllUserNotifications(ctx context.Context, req *pb.ReadAllUserNotificationsRequest) (*emptypb.Empty, error) {
	return s.readAllUserNotificationsHandler.Handle(ctx, req)
}

func (s *NotificationGRPCService) RemoveNotification(ctx context.Context, req *pb.RemoveNotificationRequest) (*emptypb.Empty, error) {
	return s.removeNotificationHandler.Handle(ctx, req)
}

func (s *NotificationGRPCService) GetUnreadCount(ctx context.Context, req *pb.GetUnreadCountRequest) (*pb.GetUnreadCountResponse, error) {
	return s.getUnreadCountHandler.Handle(ctx, req)
}
