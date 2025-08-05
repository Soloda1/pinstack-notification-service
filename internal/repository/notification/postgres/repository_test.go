package notification_repository_postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/soloda1/pinstack-proto-definitions/custom_errors"
	"pinstack-notification-service/internal/logger"
	"pinstack-notification-service/internal/model"
	notification_repository_postgres "pinstack-notification-service/internal/repository/notification/postgres"
	"pinstack-notification-service/mocks"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createSuccessCommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("INSERT 0 1")
}

func createEmptyCommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("DELETE 0")
}

func setupMockNotificationRows(t *testing.T, notifications []model.Notification) *mocks.Rows {
	mockRows := mocks.NewRows(t)
	callsCount := len(notifications)
	for i := 0; i < callsCount; i++ {
		mockRows.On("Next").Return(true).Once()
	}
	mockRows.On("Next").Return(false).Once()

	for _, notif := range notifications {
		mockRows.On("Scan",
			mock.AnythingOfType("*int64"),
			mock.AnythingOfType("*int64"),
			mock.AnythingOfType("*string"),
			mock.AnythingOfType("*bool"),
			mock.AnythingOfType("*time.Time"),
			mock.AnythingOfType("*json.RawMessage")).
			Run(func(args mock.Arguments) {
				idArg := args.Get(0).(*int64)
				userIDArg := args.Get(1).(*int64)
				typeArg := args.Get(2).(*string)
				isReadArg := args.Get(3).(*bool)
				createdAtArg := args.Get(4).(*time.Time)
				payloadArg := args.Get(5).(*json.RawMessage)

				*idArg = notif.ID
				*userIDArg = notif.UserID
				*typeArg = string(notif.Type)
				*isReadArg = notif.IsRead
				*createdAtArg = notif.CreatedAt
				*payloadArg = notif.Payload
			}).
			Return(nil).
			Once()
	}
	mockRows.On("Err").Return(nil).Maybe()
	mockRows.On("Close").Return()
	return mockRows
}

func TestNotificationRepository_Create(t *testing.T) {
	createdAt := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"event":"new_follower","follower_id":42}`)

	tests := []struct {
		name         string
		notification model.Notification
		mockSetup    func(*mocks.PgDB)
		wantErr      bool
		expectedErr  error
		expectedID   int64
	}{
		{
			name: "successful create notification",
			notification: model.Notification{
				UserID:    1,
				Type:      "relation",
				IsRead:    false,
				CreatedAt: createdAt,
				Payload:   payload,
			},
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*bool"),
					mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*json.RawMessage")).
					Run(func(args mock.Arguments) {
						idArg := args.Get(0).(*int64)
						userIDArg := args.Get(1).(*int64)
						typeArg := args.Get(2).(*string)
						isReadArg := args.Get(3).(*bool)
						createdAtArg := args.Get(4).(*time.Time)
						payloadArg := args.Get(5).(*json.RawMessage)

						*idArg = 1
						*userIDArg = 1
						*typeArg = "relation"
						*isReadArg = false
						*createdAtArg = createdAt
						*payloadArg = payload
					}).
					Return(nil)

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			wantErr:    false,
			expectedID: 1,
		},
		{
			name: "database error",
			notification: model.Notification{
				UserID:    1,
				Type:      "relation",
				IsRead:    false,
				CreatedAt: time.Now(),
				Payload:   payload,
			},
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*bool"),
					mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*json.RawMessage")).
					Return(errors.New("db error"))

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			wantErr:     true,
			expectedErr: errors.New("db error"),
			expectedID:  0,
		},
		{
			name: "postgres specific error",
			notification: model.Notification{
				UserID:    1,
				Type:      "relation",
				IsRead:    false,
				CreatedAt: time.Now(),
				Payload:   payload,
			},
			mockSetup: func(db *mocks.PgDB) {
				pgErr := &pgconn.PgError{
					Code:    "23505",
					Message: "duplicate key value violates unique constraint",
				}

				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*bool"),
					mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*json.RawMessage")).
					Return(pgErr)

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
			expectedID:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := notification_repository_postgres.NewNotificationRepository(mockDB, log)
			id, err := repo.Create(context.Background(), &tt.notification)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), id)
				if errors.Is(err, custom_errors.ErrDatabaseQuery) {
					assert.ErrorIs(t, err, custom_errors.ErrDatabaseQuery)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
				assert.Equal(t, tt.expectedID, tt.notification.ID)
			}
		})
	}
}

func TestNotificationRepository_GetByID(t *testing.T) {
	createdAt := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	payload := json.RawMessage(`{"event":"new_follower","follower_id":42}`)

	tests := []struct {
		name        string
		id          int64
		mockSetup   func(*mocks.PgDB)
		want        *model.Notification
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful get notification",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*bool"),
					mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*json.RawMessage")).
					Run(func(args mock.Arguments) {
						idArg := args.Get(0).(*int64)
						userIDArg := args.Get(1).(*int64)
						typeArg := args.Get(2).(*string)
						isReadArg := args.Get(3).(*bool)
						createdAtArg := args.Get(4).(*time.Time)
						payloadArg := args.Get(5).(*json.RawMessage)

						*idArg = 1
						*userIDArg = 2
						*typeArg = "relation"
						*isReadArg = false
						*createdAtArg = createdAt
						*payloadArg = payload
					}).
					Return(nil)

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			want: &model.Notification{
				ID:        1,
				UserID:    2,
				Type:      "relation",
				IsRead:    false,
				CreatedAt: createdAt,
				Payload:   payload,
			},
			wantErr: false,
		},
		{
			name: "notification not found",
			id:   999,
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*bool"),
					mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*json.RawMessage")).
					Return(pgx.ErrNoRows)

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrNotificationNotFound,
		},
		{
			name: "database error",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*bool"),
					mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*json.RawMessage")).
					Return(errors.New("db error"))

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "postgres specific error",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				pgErr := &pgconn.PgError{
					Code:    "42P01",
					Message: "relation \"notifications\" does not exist",
				}

				mockRow := new(mocks.Row)
				mockRow.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*bool"),
					mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*json.RawMessage")).
					Return(pgErr)

				db.On("QueryRow",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRow)
			},
			want:        nil,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := notification_repository_postgres.NewNotificationRepository(mockDB, log)
			got, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.UserID, got.UserID)
				assert.Equal(t, tt.want.Type, got.Type)
				assert.Equal(t, tt.want.IsRead, got.IsRead)
				assert.Equal(t, tt.want.CreatedAt, got.CreatedAt)
				assert.Equal(t, string(tt.want.Payload), string(got.Payload))
			}
		})
	}
}

func TestNotificationRepository_ListByUser(t *testing.T) {
	createdAt1 := time.Date(2025, 6, 16, 12, 0, 0, 0, time.UTC)
	createdAt2 := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	payload1 := json.RawMessage(`{"event":"new_follower","follower_id":42}`)
	payload2 := json.RawMessage(`{"event":"new_follower","follower_id":43}`)

	notif1 := model.Notification{
		ID:        1,
		UserID:    5,
		Type:      "relation",
		IsRead:    false,
		CreatedAt: createdAt1,
		Payload:   payload1,
	}

	notif2 := model.Notification{
		ID:        2,
		UserID:    5,
		Type:      "relation",
		IsRead:    true,
		CreatedAt: createdAt2,
		Payload:   payload2,
	}

	tests := []struct {
		name        string
		userID      int64
		limit       int
		offset      int
		mockSetup   func(*mocks.PgDB)
		want        []*model.Notification
		wantTotal   int32
		wantErr     bool
		expectedErr error
		checkQuery  bool
	}{
		{
			name:   "successful get notifications with default pagination",
			userID: 5,
			limit:  10,
			offset: 0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int32")).
					Run(func(args mock.Arguments) {
						totalPtr := args[0].(*int32)
						*totalPtr = 2
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "COUNT(*)")
					}),
					mock.Anything).Return(mockCountRow)

				// Mock list query
				rows := setupMockNotificationRows(t, []model.Notification{notif1, notif2})
				db.On("Query",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "SELECT id, user_id")
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["user_id"] == int64(5) &&
							args["limit"] == 10 &&
							args["offset"] == 0
					})).Return(rows, nil)
			},
			want:       []*model.Notification{&notif1, &notif2},
			wantTotal:  2,
			wantErr:    false,
			checkQuery: true,
		},
		{
			name:   "get notifications with custom pagination",
			userID: 5,
			limit:  5,
			offset: 10,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int32")).
					Run(func(args mock.Arguments) {
						totalPtr := args[0].(*int32)
						*totalPtr = 2
					}).Return(nil)
				db.On("QueryRow",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "COUNT(*)")
					}),
					mock.Anything).Return(mockCountRow)

				// Mock list query
				rows := setupMockNotificationRows(t, []model.Notification{notif2})
				db.On("Query",
					mock.Anything,
					mock.MatchedBy(func(query string) bool {
						return strings.Contains(query, "SELECT id, user_id")
					}),
					mock.MatchedBy(func(args pgx.NamedArgs) bool {
						return args["user_id"] == int64(5) &&
							args["limit"] == 5 &&
							args["offset"] == 10
					})).Return(rows, nil)
			},
			want:       []*model.Notification{&notif2},
			wantTotal:  2,
			wantErr:    false,
			checkQuery: true,
		},
		{
			name:   "empty notifications list",
			userID: 5,
			limit:  10,
			offset: 0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int32")).
					Run(func(args mock.Arguments) {
						totalPtr := args[0].(*int32)
						*totalPtr = 0
					}).Return(nil)
				db.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(mockCountRow)

				// Mock list query
				rows := setupMockNotificationRows(t, []model.Notification{})
				db.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(rows, nil)
			},
			want:      []*model.Notification{},
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:   "database query error",
			userID: 5,
			limit:  10,
			offset: 0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query error
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int32")).Return(errors.New("db error"))
				db.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(mockCountRow)
			},
			want:        nil,
			wantTotal:   0,
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
		{
			name:   "scan error",
			userID: 5,
			limit:  10,
			offset: 0,
			mockSetup: func(db *mocks.PgDB) {
				// Mock count query success
				mockCountRow := new(mocks.Row)
				mockCountRow.On("Scan", mock.AnythingOfType("*int32")).
					Run(func(args mock.Arguments) {
						totalPtr := args[0].(*int32)
						*totalPtr = 1
					}).Return(nil)
				db.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(mockCountRow)

				// Mock list query with scan error
				mockRows := mocks.NewRows(t)
				mockRows.On("Next").Return(true).Once()
				mockRows.On("Scan",
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*int64"),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType("*bool"),
					mock.AnythingOfType("*time.Time"),
					mock.AnythingOfType("*json.RawMessage")).Return(errors.New("scan error"))
				mockRows.On("Close").Return()
				db.On("Query",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(mockRows, nil)
			},
			want:        nil,
			wantTotal:   0,
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := notification_repository_postgres.NewNotificationRepository(mockDB, log)
			got, gotTotal, err := repo.ListByUser(context.Background(), tt.userID, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantTotal, gotTotal)
				if tt.expectedErr != nil && errors.Is(tt.expectedErr, custom_errors.ErrDatabaseQuery) {
					assert.ErrorIs(t, err, custom_errors.ErrDatabaseQuery)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantTotal, gotTotal)
				assert.Equal(t, len(tt.want), len(got))

				for i, wantNotif := range tt.want {
					assert.Equal(t, wantNotif.ID, got[i].ID)
					assert.Equal(t, wantNotif.UserID, got[i].UserID)
					assert.Equal(t, wantNotif.Type, got[i].Type)
					assert.Equal(t, wantNotif.IsRead, got[i].IsRead)
					assert.Equal(t, wantNotif.CreatedAt, got[i].CreatedAt)
					assert.Equal(t, string(wantNotif.Payload), string(got[i].Payload))
				}
			}
		})
	}
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		mockSetup   func(*mocks.PgDB)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful mark as read",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(createSuccessCommandTag(), nil)
			},
			wantErr: false,
		},
		{
			name: "notification not found",
			id:   999,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(createEmptyCommandTag(), nil)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrNotificationNotFound,
		},
		{
			name: "database error",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(pgconn.CommandTag{}, errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
		{
			name: "postgres specific error",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				pgErr := &pgconn.PgError{
					Code:    "42P01",
					Message: "relation \"notifications\" does not exist",
				}
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(pgconn.CommandTag{}, pgErr)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := notification_repository_postgres.NewNotificationRepository(mockDB, log)
			err := repo.MarkAsRead(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					if errors.Is(err, custom_errors.ErrNotificationNotFound) || errors.Is(err, custom_errors.ErrDatabaseQuery) {
						assert.ErrorIs(t, err, tt.expectedErr)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotificationRepository_MarkAllAsRead(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		mockSetup   func(*mocks.PgDB)
		wantErr     bool
		expectedErr error
	}{
		{
			name:   "successful mark all as read",
			userID: 5,
			mockSetup: func(db *mocks.PgDB) {
				// Simulate 3 rows affected (3 notifications marked as read)
				commandTag := pgconn.NewCommandTag("UPDATE 3")
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(commandTag, nil)
			},
			wantErr: false,
		},
		{
			name:   "no notifications to mark as read",
			userID: 5,
			mockSetup: func(db *mocks.PgDB) {
				// No rows affected is not an error for MarkAllAsRead
				commandTag := pgconn.NewCommandTag("UPDATE 0")
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(commandTag, nil)
			},
			wantErr: false,
		},
		{
			name:   "database error",
			userID: 5,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(pgconn.CommandTag{}, errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
		{
			name:   "postgres specific error",
			userID: 5,
			mockSetup: func(db *mocks.PgDB) {
				pgErr := &pgconn.PgError{
					Code:    "42P01",
					Message: "relation \"notifications\" does not exist",
				}
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(pgconn.CommandTag{}, pgErr)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := notification_repository_postgres.NewNotificationRepository(mockDB, log)
			err := repo.MarkAllAsRead(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil && errors.Is(err, custom_errors.ErrDatabaseQuery) {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNotificationRepository_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		mockSetup   func(*mocks.PgDB)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful delete",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(createSuccessCommandTag(), nil)
			},
			wantErr: false,
		},
		{
			name: "notification not found",
			id:   999,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(createEmptyCommandTag(), nil)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrNotificationNotFound,
		},
		{
			name: "database error",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(pgconn.CommandTag{}, errors.New("db error"))
			},
			wantErr:     true,
			expectedErr: errors.New("db error"),
		},
		{
			name: "postgres specific error",
			id:   1,
			mockSetup: func(db *mocks.PgDB) {
				pgErr := &pgconn.PgError{
					Code:    "42P01",
					Message: "relation \"notifications\" does not exist",
				}
				db.On("Exec",
					mock.Anything,
					mock.AnythingOfType("string"),
					mock.Anything).Return(pgconn.CommandTag{}, pgErr)
			},
			wantErr:     true,
			expectedErr: custom_errors.ErrDatabaseQuery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := mocks.NewPgDB(t)
			log := logger.New("dev")

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			repo := notification_repository_postgres.NewNotificationRepository(mockDB, log)
			err := repo.Delete(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					if errors.Is(err, custom_errors.ErrNotificationNotFound) || errors.Is(err, custom_errors.ErrDatabaseQuery) {
						assert.ErrorIs(t, err, tt.expectedErr)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
