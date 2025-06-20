package notification_repository_postgres

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"log/slog"
	"pinstack-notification-service/internal/custom_errors"
	"pinstack-notification-service/internal/logger"
	"pinstack-notification-service/internal/model"
	"pinstack-notification-service/internal/repository/db"
	"time"
)

type NotificationRepository struct {
	log *logger.Logger
	db  db.PgDB
}

func NewNotificationRepository(db db.PgDB, log *logger.Logger) *NotificationRepository {
	return &NotificationRepository{db: db, log: log}
}

func (r *NotificationRepository) Create(ctx context.Context, notif *model.Notification) error {
	createdAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if !notif.CreatedAt.IsZero() {
		createdAt.Time = notif.CreatedAt
	}

	args := pgx.NamedArgs{
		"user_id":    notif.UserID,
		"type":       notif.Type,
		"is_read":    notif.IsRead,
		"created_at": createdAt,
		"payload":    notif.Payload,
	}

	query := `
		INSERT INTO notifications (
			user_id, 
			type, 
			is_read, 
			created_at, 
			payload
		) VALUES (
			@user_id, 
			@type, 
			@is_read, 
			@created_at, 
			@payload
		) RETURNING id, user_id, type, is_read, created_at, payload
	`

	r.log.Debug("Creating notification",
		slog.Int64("user_id", notif.UserID),
		slog.String("type", notif.Type),
	)

	var createdNotification model.Notification
	err := r.db.QueryRow(ctx, query, args).Scan(
		&createdNotification.ID,
		&createdNotification.UserID,
		&createdNotification.Type,
		&createdNotification.IsRead,
		&createdNotification.CreatedAt,
		&createdNotification.Payload,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to create notification",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("user_id", notif.UserID),
			)

			return custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to create notification", slog.String("error", err.Error()))
		return err
	}

	r.log.Debug("Notification created successfully",
		slog.Int64("id", createdNotification.ID),
		slog.Int64("user_id", createdNotification.UserID),
	)

	return nil
}
