package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// работа с таблицей subscriptions
type SubscriptionRepository struct {
	db *sql.DB
}

// NewSubscriptionRepository создает новый репозиторий подписок
func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Upsert создание или обновление подписки
func (r *SubscriptionRepository) Upsert(ctx context.Context, sub model.Subscription) error {
	query := `
INSERT INTO subscriptions (chat_id, cryptocurrency, interval_min, is_active, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (chat_id, cryptocurrency)
DO UPDATE SET
interval_min = $3,
is_active = $4,
updated_at = NOW()
`

	_, err := r.db.ExecContext(ctx, query,
		sub.ChatID,
		sub.Cryptocurrency,
		sub.IntervalMin,
		sub.IsActive,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert subscription: %w", err)
	}
	return nil
}

// Deactive отключает подписку
func (r *SubscriptionRepository) Deactivate(ctx context.Context, chatID int64, cryptocurrency string) error {
	query := `
UPDATE subscriptions
SET is_active = FALSE, updated_at = NOW()
WHERE chat_id = $1 AND cryptocurrency = $2
`

	_, err := r.db.ExecContext(ctx, query, chatID, cryptocurrency)
	if err != nil {
		return fmt.Errorf("failed to deactivate subscription: %w", err)
	}
	return nil
}

// GetActive возвратит активные подписки
func (r *SubscriptionRepository) GetActive(ctx context.Context) ([]model.Subscription, error) {
	query := `
SELECT id, chat_id< cryptocurrency, interval_min, is_active, created_at, updated_at
FROM subscriptions
WHERE is_active = TRUE
ORDER BY chat_id, cryptocurrency
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions : %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ChatID,
			&sub.Cryptocurrency,
			&sub.IntervalMin,
			&sub.IsActive,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

// GetByChatID возвращает любые подписки конкретного чата
func (r *SubscriptionRepository) GetByChatID(ctx context.Context, chatID int64) ([]model.Subscription, error) {
	query := `
SELECT id, chat_id, cryptocurrency, interval_min, is_active, created_at, updated_at
FROM subscriptions
WHERE chat_id = $1
ORDER BY cryptocurrency
`

	rows, err := r.db.QueryContext(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ChatID,
			&sub.Cryptocurrency,
			&sub.IntervalMin,
			&sub.IsActive,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failedd to scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

// CleanupOld пока не используется, создается для будущей очистки старых записей
func (r *SubscriptionRepository) CleanupOld(ctx context.Context, before time.Time) error {
	query := `DELETE FROM subscriptions WHERE is_active = FALSE AND updated_at < $1`
	_, err := r.db.ExecContext(ctx, query, before)
	return err
}

// GetDueSubscription возвращает подписки, которые пора отправить
// если подписка на 10 минут, отправляем каждые 10 минут от времени создания
func (r *SubscriptionRepository) GetDueSubscriptions(ctx context.Context) ([]model.Subscription, error) {
	query := `
		SELECT id, chat_id, cryptocurrency, interval_min, is_active, created_at, updated_at
		FROM subscriptions
		WHERE is_active = TRUE
		AND (
			-- Пора отправлять: прошло достаточно минут с последнего обновления
			updated_at + (interval_min || ' minutes')::INTERVAL <= NOW()
		)
		ORDER BY chat_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get due subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ChatID,
			&sub.Cryptocurrency,
			&sub.IntervalMin,
			&sub.IsActive,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}

	return subs, rows.Err()
}

// Touch обновляет updated_at для подписки (чтобы не отправлять повторно)
func (r *SubscriptionRepository) Touch(ctx context.Context, id int64) error {
	query := `UPDATE subscriptions SET updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to touch subscription: %w", err)
	}
	return nil
}
