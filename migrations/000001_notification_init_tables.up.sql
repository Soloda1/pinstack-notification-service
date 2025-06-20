CREATE TABLE notifications (
   id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
   user_id bigint NOT NULL,
   type TEXT NOT NULL,
   is_read BOOLEAN NOT NULL DEFAULT FALSE,
   created_at TIMESTAMP NOT NULL DEFAULT NOW(),
   payload JSONB
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);