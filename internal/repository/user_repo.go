package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// UserRepository - работа с таблицей users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository создание нового репо пользователей
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Upsert создание или обновление пользователя при активности
func (r *UserRepository) Upsert(ctx context.Context, user model.User) error {
	query := `
		INSERT INTO users (chat_id, username, first_name, last_name, language_code, first_seen, last_active, command_count)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), 1)
		ON CONFLICT (chat_id) 
		DO UPDATE SET 
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			language_code = EXCLUDED.language_code,
			last_active = NOW(),
			command_count = users.command_count + 1
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ChatID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.LanguageCode,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}
	return nil
}

// GetByChatID возвращает пользователя по chat_id
func (r *UserRepository) GetByChatID(ctx context.Context, chatID int64) (*model.User, error) {
	query := `
SELECT id, chat_id, username, first_name, last_name, language_code, first_seen, last_active, command_count
FROM users
WHERE chat_id = $1
`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(
		&user.ID,
		&user.ChatID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.LanguageCode,
		&user.FirstSeen,
		&user.LastActive,
		&user.CommandCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetStats возврат общей статистики по пользакам
func (r *UserRepository) GetStats(ctx context.Context) (*UserStats, error) {
	query := `
SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN last_active >= NOW() - INTERVAL '24 hours' THEN 1 END) as active_24h,
			COUNT(CASE WHEN last_active >= NOW() - INTERVAL '7 days' THEN 1 END) as active_7d,
			COALESCE(SUM(command_count), 0) as total_commands
		FROM users
`

	var stats UserStats
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.Active24h,
		&stats.Active7d,
		&stats.TotalCommands,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &stats, nil
}

// UserStats - статистика по пользователю
type UserStats struct {
	TotalUsers    int `json:"total_users"`
	Active24h     int `json:"active_24h"`
	Active7d      int `json:"active_7d"`
	TotalCommands int `json:"total_commands"`
}

// emptyToNull заменяет пустую строку на NULL
func emptyToNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// GetAllUsers возвращает всех пользователей, отсортированных по последней активности
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]model.User, error) {
	query := `
		SELECT id, chat_id, username, first_name, last_name, language_code, first_seen, last_active, command_count
		FROM users
		ORDER BY last_active DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(
			&u.ID, &u.ChatID, &u.Username, &u.FirstName, &u.LastName,
			&u.LanguageCode, &u.FirstSeen, &u.LastActive, &u.CommandCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, rows.Err()
}
