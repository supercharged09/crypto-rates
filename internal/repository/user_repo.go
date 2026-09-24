package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// UserRepository - работа с таблицей users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository создание нового репо пользователей
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Upsert создание или обновление пользователя при активности
func (r *UserRepository) Upsert(ctx context.Context, user model.User) error {
	user.FirstSeen = time.Now()
	user.LastActive = time.Now()
	user.CommandCount = 1

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "chat_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"username":      user.Username,
				"first_name":    user.FirstName,
				"last_name":     user.LastName,
				"language_code": user.LanguageCode,
				"last_active":   time.Now(),
				"command_count": gorm.Expr("users.command_count + 1"),
			}),
		}).
		Create(&user).Error

	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}
	return nil
}

// GetByChatID возвращает пользователя по chat_id
func (r *UserRepository) GetByChatID(ctx context.Context, chatID int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetStats возврат общей статистики по пользователям
func (r *UserRepository) GetStats(ctx context.Context) (*UserStats, error) {
	var stats UserStats

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Select(`
			COUNT(*) as total_users,
			COUNT(CASE WHEN last_active >= NOW() - INTERVAL '24 hours' THEN 1 END) as active_24h,
			COUNT(CASE WHEN last_active >= NOW() - INTERVAL '7 days' THEN 1 END) as active_7d,
			COALESCE(SUM(command_count), 0) as total_commands
		`).
		Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return &stats, nil
}

// UserStats - статистика по пользователям
type UserStats struct {
	TotalUsers    int `json:"total_users"`
	Active24h     int `json:"active_24h"`
	Active7d      int `json:"active_7d"`
	TotalCommands int `json:"total_commands"`
}

// GetAllUsers возвращает всех пользователей
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Order("last_active DESC").
		Find(&users).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
}
