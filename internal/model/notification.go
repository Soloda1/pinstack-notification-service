package model

type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt int64                  `json:"created_at"`
	Read      bool                   `json:"read"`
}
