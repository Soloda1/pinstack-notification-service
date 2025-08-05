package notification_repository_postgres

import (
	"context"
	"errors"
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"log/slog"
	"pinstack-notification-service/internal/logger"
	"pinstack-notification-service/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/soloda1/pinstack-proto-definitions/events"
)

type NotificationRepository struct {
	log *logger.Logger
	db  PgDB
}

func NewNotificationRepository(db PgDB, log *logger.Logger) *NotificationRepository {
	return &NotificationRepository{db: db, log: log}
}

func (r *NotificationRepository) Create(ctx context.Context, notif *model.Notification) (int64, error) {
	createdAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	if !notif.CreatedAt.IsZero() {
		createdAt.Time = notif.CreatedAt
	}

	args := pgx.NamedArgs{
		"user_id":    notif.UserID,
		"type":       string(notif.Type),
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
		slog.String("type", string(notif.Type)),
	)

	var createdNotification model.Notification
	var typeStr string
	err := r.db.QueryRow(ctx, query, args).Scan(
		&createdNotification.ID,
		&createdNotification.UserID,
		&typeStr,
		&createdNotification.IsRead,
		&createdNotification.CreatedAt,
		&createdNotification.Payload,
	)
	createdNotification.Type = events.EventType(typeStr)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to create notification",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("user_id", notif.UserID),
			)

			return 0, custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to create notification", slog.String("error", err.Error()))
		return 0, err
	}

	// Update the passed notification object with the created data
	notif.ID = createdNotification.ID
	notif.CreatedAt = createdNotification.CreatedAt

	r.log.Debug("Notification created successfully",
		slog.Int64("id", createdNotification.ID),
		slog.Int64("user_id", createdNotification.UserID),
	)

	return createdNotification.ID, nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id int64) (*model.Notification, error) {
	query := `
		SELECT id, user_id, type, is_read, created_at, payload 
		FROM notifications 
		WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": id,
	}

	r.log.Debug("Getting notification by ID", slog.Int64("id", id))

	var notification model.Notification
	var typeStr string
	err := r.db.QueryRow(ctx, query, args).Scan(
		&notification.ID,
		&notification.UserID,
		&typeStr,
		&notification.IsRead,
		&notification.CreatedAt,
		&notification.Payload,
	)
	notification.Type = events.EventType(typeStr)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Debug("Notification not found", slog.Int64("id", id))
			return nil, custom_errors.ErrNotificationNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to get notification",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("id", id),
			)

			return nil, custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to get notification", slog.String("error", err.Error()))
		return nil, err
	}

	r.log.Debug("Notification retrieved successfully",
		slog.Int64("id", notification.ID),
		slog.Int64("user_id", notification.UserID),
	)

	return &notification, nil
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID int64, limit int, offset int) ([]*model.Notification, int32, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM notifications 
		WHERE user_id = @user_id
	`

	countArgs := pgx.NamedArgs{
		"user_id": userID,
	}

	var totalCount int32
	err := r.db.QueryRow(ctx, countQuery, countArgs).Scan(&totalCount)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to count notifications",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("user_id", userID),
			)

			return nil, 0, custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to count notifications", slog.String("error", err.Error()))
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, type, is_read, created_at, payload 
		FROM notifications 
		WHERE user_id = @user_id
		ORDER BY created_at DESC
		LIMIT @limit OFFSET @offset
	`

	args := pgx.NamedArgs{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}

	r.log.Debug("Listing notifications by user",
		slog.Int64("user_id", userID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
		slog.Int("total_count", int(totalCount)),
	)

	rows, err := r.db.Query(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to list notifications",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("user_id", userID),
			)

			return nil, 0, custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to list notifications", slog.String("error", err.Error()))
		return nil, 0, err
	}
	defer rows.Close()

	notifications := make([]*model.Notification, 0)
	for rows.Next() {
		var notification model.Notification
		var typeStr string
		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&typeStr,
			&notification.IsRead,
			&notification.CreatedAt,
			&notification.Payload,
		)
		notification.Type = events.EventType(typeStr)

		if err != nil {
			r.log.Error("Failed to scan notification row", slog.String("error", err.Error()))
			return nil, 0, custom_errors.ErrDatabaseQuery
		}

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		r.log.Error("Error during rows iteration", slog.String("error", err.Error()))
		return nil, 0, err
	}

	r.log.Debug("Retrieved notifications successfully",
		slog.Int64("user_id", userID),
		slog.Int("count", len(notifications)),
		slog.Int("total_count", int(totalCount)),
	)

	return notifications, totalCount, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id int64) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": id,
	}

	r.log.Debug("Marking notification as read", slog.Int64("id", id))

	result, err := r.db.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to mark notification as read",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("id", id),
			)

			return custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to mark notification as read", slog.String("error", err.Error()))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.log.Debug("Notification not found", slog.Int64("id", id))
		return custom_errors.ErrNotificationNotFound
	}

	r.log.Debug("Notification marked as read successfully", slog.Int64("id", id))
	return nil
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	query := `
		UPDATE notifications
		SET is_read = true
		WHERE user_id = @user_id AND is_read = false
	`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	r.log.Debug("Marking all notifications as read for user", slog.Int64("user_id", userID))

	result, err := r.db.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to mark all notifications as read",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("user_id", userID),
			)

			return custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to mark all notifications as read", slog.String("error", err.Error()))
		return err
	}

	rowsAffected := result.RowsAffected()
	r.log.Debug("Notifications marked as read successfully",
		slog.Int64("user_id", userID),
		slog.Int64("count", rowsAffected),
	)

	return nil
}

func (r *NotificationRepository) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM notifications
		WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": id,
	}

	r.log.Debug("Deleting notification", slog.Int64("id", id))

	result, err := r.db.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to delete notification",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("id", id),
			)

			return custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to delete notification", slog.String("error", err.Error()))
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.log.Debug("Notification not found", slog.Int64("id", id))
		return custom_errors.ErrNotificationNotFound
	}

	r.log.Debug("Notification deleted successfully", slog.Int64("id", id))
	return nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = @user_id AND is_read = false
	`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	r.log.Debug("Counting unread notifications for user", slog.Int64("user_id", userID))

	var count int
	err := r.db.QueryRow(ctx, query, args).Scan(&count)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.log.Error("Failed to count unread notifications",
				slog.String("pg_error_code", pgErr.Code),
				slog.String("pg_error_message", pgErr.Message),
				slog.String("pg_error_detail", pgErr.Detail),
				slog.Int64("user_id", userID),
			)

			return 0, custom_errors.ErrDatabaseQuery
		}

		r.log.Error("Failed to count unread notifications", slog.String("error", err.Error()))
		return 0, err
	}

	r.log.Debug("Counted unread notifications successfully",
		slog.Int64("user_id", userID),
		slog.Int("count", count),
	)

	return count, nil
}
