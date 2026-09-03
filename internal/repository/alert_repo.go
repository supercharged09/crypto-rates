package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// AlertRepository — работа с таблицей alerts
type AlertRepository struct {
	db *sql.DB
}

// NewAlertRepository создаёт новый репозиторий алертов
func NewAlertRepository(db *sql.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Create создаёт новый алерт
func (r *AlertRepository) Create(ctx context.Context, alert model.Alert) error {
	query := `
		INSERT INTO alerts (chat_id, cryptocurrency, direction, price_threshold, is_active, created_at)
		VALUES ($1, $2, $3, $4, TRUE, NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		alert.ChatID,
		alert.Cryptocurrency,
		alert.Direction,
		alert.PriceThreshold,
	)
	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetActive возвращает все активные алерты
func (r *AlertRepository) GetActive(ctx context.Context) ([]model.Alert, error) {
	query := `
		SELECT id, chat_id, cryptocurrency, direction, price_threshold, is_active, created_at, triggered_at
		FROM alerts
		WHERE is_active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		var a model.Alert
		err := rows.Scan(
			&a.ID,
			&a.ChatID,
			&a.Cryptocurrency,
			&a.Direction,
			&a.PriceThreshold,
			&a.IsActive,
			&a.CreatedAt,
			&a.TriggeredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	return alerts, rows.Err()
}

// GetByChatID возвращает алерты пользователя
func (r *AlertRepository) GetByChatID(ctx context.Context, chatID int64) ([]model.Alert, error) {
	query := `
		SELECT id, chat_id, cryptocurrency, direction, price_threshold, is_active, created_at, triggered_at
		FROM alerts
		WHERE chat_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		var a model.Alert
		err := rows.Scan(
			&a.ID,
			&a.ChatID,
			&a.Cryptocurrency,
			&a.Direction,
			&a.PriceThreshold,
			&a.IsActive,
			&a.CreatedAt,
			&a.TriggeredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}

	return alerts, rows.Err()
}

// Deactivate отключает алерт
func (r *AlertRepository) Deactivate(ctx context.Context, id int64) error {
	query := `UPDATE alerts SET is_active = FALSE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate alert: %w", err)
	}
	return nil
}

// MarkTriggered помечает алерт как сработавший
func (r *AlertRepository) MarkTriggered(ctx context.Context, id int64) error {
	query := `UPDATE alerts SET is_active = FALSE, triggered_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark alert triggered: %w", err)
	}
	return nil
}
