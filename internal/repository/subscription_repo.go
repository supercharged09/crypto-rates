package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// работа с таблицей subscriptions
type SubscriptionRepository struct {
	db *gorm.DB
}

// создание нового репо подписок
func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// создание или обновление подписки
func (r *SubscriptionRepository) Upsert(ctx context.Context, sub model.Subscription) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "chat_id"}, {Name: "cryptocurrency"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"interval_min": sub.IntervalMin,
				"is_active":    true,
				"updated_at":   time.Now(),
			}),
		}).
		Create(&sub).Error

	if err != nil {
		return fmt.Errorf("failed to upsert subscription: %w", err)
	}
	return nil
}

// отключение подписки
func (r *SubscriptionRepository) Deactivate(ctx context.Context, chatID int64, cryptocurrency string) error {
	err := r.db.WithContext(ctx).
		Model(&model.Subscription{}).
		Where("chat_id = ? AND cryptocurrency = ?", chatID, cryptocurrency).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to deactivate sunscription: %w", err)
	}
	return nil
}

// GetDueSubscriptions возвращает подписки, готовые к отправке
func (r *SubscriptionRepository) GetDueSubscriptions(ctx context.Context) ([]model.Subscription, error) {
	var subs []model.Subscription

	//поиск подписок, где updated at + interval min <= NOW()
	err := r.db.WithContext(ctx).
		Where("is_active = TRUE").
		Where("updated_at + (inteval_min * INTERVAL '1 minute') <= ?", time.Now()).
		Order("chat_id").
		Find(&subs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get due subscriptions: %w", err)
	}
	return subs, nil
}

// Touch обновляет updated at для подписки
func (r *SubscriptionRepository) Touch(ctx context.Context, id int64) error {
	err := r.db.WithContext(ctx).
		Model(&model.Subscription{}).
		Where("id = ?", id).
		Update("updated_at", time.Now()).Error

	if err != nil {
		return fmt.Errorf("failed to touch subscription: %w", err)
	}
	return nil
}

// GetActive для возвращения всех активных подписок
func (r *SubscriptionRepository) GetActive(ctx context.Context) ([]model.Subscription, error) {
	var subs []model.Subscription
	err := r.db.WithContext(ctx).
		Where("is_active = TRUE").
		Order("chat_id, cryptocurrency").
		Find(&subs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}
	return subs, nil
}

// GetByChatID вернет все подписки конкретного чата
func (r *SubscriptionRepository) GetByChatID(ctx context.Context, chatID int64) ([]model.Subscription, error) {
	var subs []model.Subscription
	err := r.db.WithContext(ctx).
		Where("chat_id = ?", chatID).
		Order("cryptocurrency").
		Find(&subs).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}
	return subs, nil
}

// CleanupOld удалит старые неактивные подписки
func (r *SubscriptionRepository) CleanupOld(ctx context.Context, before time.Time) error {
	err := r.db.WithContext(ctx).
		Where("is-active = FALSE AND updated_at < ?", before).
		Delete(&model.Subscription{}).Error

	if err != nil {
		return fmt.Errorf("failed to cleanup subscriptions: %w", err)
	}
	return nil
}
