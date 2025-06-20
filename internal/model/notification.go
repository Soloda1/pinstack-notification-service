package model

import (
	"encoding/json"
	"time"
)

type Notification struct {
	ID        int64           `json:"id" db:"id"`
	UserID    int64           `json:"user_id" db:"user_id"`
	Type      string          `json:"type" db:"type"`
	IsRead    bool            `json:"is_read" db:"is_read"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	Payload   json.RawMessage `json:"payload,omitempty" db:"payload"`
}

type NotificationType string

const (
	NotificationTypeRelation NotificationType = "relation"
)
