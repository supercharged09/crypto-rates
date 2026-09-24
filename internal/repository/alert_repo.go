package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// AlertRepository — работа с таблицей alerts
type AlertRepository struct {
	db *gorm.DB
}

// NewAlertRepository создаёт новый репозиторий алертов
func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Create создаёт новый алерт
func (r *AlertRepository) Create(ctx context.Context, alert model.Alert) error {
	if err := r.db.WithContext(ctx).Create(&alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}
	return nil
}

// GetActive возвращает все активные алерты
func (r *AlertRepository) GetActive(ctx context.Context) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.WithContext(ctx).
		Where("is_active = TRUE").
		Order("created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}
	return alerts, nil
}

// GetByChatID возвращает алерты пользователя
func (r *AlertRepository) GetByChatID(ctx context.Context, chatID int64) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	return alerts, nil
}

// Deactivate отключает алерт
func (r *AlertRepository) Deactivate(ctx context.Context, id int64) error {
	err := r.db.WithContext(ctx).
		Model(&model.Alert{}).
		Where("id = ?", id).
		Update("is_active", false).Error

	if err != nil {
		return fmt.Errorf("failed to deactivate alert: %w", err)
	}
	return nil
}

// MarkTriggered помечает алерт как сработавший
func (r *AlertRepository) MarkTriggered(ctx context.Context, id int64) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&model.Alert{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_active":    false,
			"triggered_at": now,
		}).Error

	if err != nil {
		return fmt.Errorf("failed to mark alert triggered: %w", err)
	}
	return nil
}
