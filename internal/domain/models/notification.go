package models

import (
	"encoding/json"
	"time"

	"github.com/soloda1/pinstack-proto-definitions/events"
)

type Notification struct {
	ID        int64            `json:"id" db:"id"`
	UserID    int64            `json:"user_id" db:"user_id"`
	Type      events.EventType `json:"type" db:"type"`
	IsRead    bool             `json:"is_read" db:"is_read"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	Payload   json.RawMessage  `json:"payload,omitempty" db:"payload"`
}
