package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// RateRepositoryInterface интерфейс для тестирования
type RateRepositoryInterface interface {
	Save(ctx context.Context, rate model.Rate) error
	GetCurrentPrice(ctx context.Context, cryptocurrency string) (*model.Rate, error)
	GetMinMax24h(ctx context.Context, cryptocurrency string) (float64, float64, error)
	GetPriceHourAgo(ctx context.Context, cryptocurrency string) (float64, error)
	GetPrices24h(ctx context.Context, cryptocurrency string) ([]model.PricePoint, error)
}

// работа с таблицей через горм
type RateRepository struct {
	db *gorm.DB
}

// NewRateRepository создание нового репо
func NewRateRepository(db *gorm.DB) *RateRepository {
	return &RateRepository{db: db}
}

// Сохранение курса в БД
func (r *RateRepository) Save(ctx context.Context, rate model.Rate) error {
	if err := r.db.WithContext(ctx).Create(&rate).Error; err != nil {
		return fmt.Errorf("failed to save rate: %w", err)
	}
	return nil
}

// GetCurrentPrice получение последней цены для крипты
func (r *RateRepository) GetCurrentPrice(ctx context.Context, cryptocurrency string) (*model.Rate, error) {
	var rate model.Rate
	err := r.db.WithContext(ctx).
		Where("cryptocurrency = ?", cryptocurrency).
		Order("timestamp DESC").
		First(&rate).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current price: %w", err)
	}

	return &rate, nil
}

// Получение минимальной и максимальной цены за 24 часа
func (r *RateRepository) GetMinMax24h(ctx context.Context, cryptocurrency string) (float64, float64, error) {
	var result struct {
		MinCents *int64
		MaxCents *int64
	}

	err := r.db.WithContext(ctx).
		Model(&model.Rate{}).
		Select("MIN(price_usd_cents)as min_cents, MAX(price_usd_cents) as max_cents").
		Where("cryptocurrency = ? AND timestamp >= ?", cryptocurrency, time.Now().Add(-24*time.Hour)).
		Scan(&result).Error

	if err != nil {
		return 0, 0, fmt.Errorf("failed to get min/max: %w", err)
	}

	min := float64(0)
	max := float64(0)
	if result.MinCents != nil {
		min = model.CentsToFloat(*result.MinCents)
	}
	if result.MaxCents != nil {
		max = model.CentsToFloat(*result.MaxCents)
	}

	return min, max, nil
}

// Получение цены за час назад
func (r *RateRepository) GetPriceHourAgo(ctx context.Context, cryptocurrency string) (float64, error) {
	var rate model.Rate
	err := r.db.WithContext(ctx).
		Where("cryptocurrency = ? AND timestamp <= ?", cryptocurrency, time.Now().Add(-1*time.Hour)).
		Order("timestamp DESC").
		First(&rate).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get price hour ago: %w", err)
	}

	return model.CentsToFloat(rate.PriceUSDCents), nil
}

// Возврат цены за последние 24 часа
func (r *RateRepository) GetPrices24h(ctx context.Context, cryptocurrency string) ([]model.PricePoint, error) {
	var rates []model.Rate
	err := r.db.WithContext(ctx).
		Where("cryptocurrency = ? AND timestamp >= ?", cryptocurrency, time.Now().Add(-24*time.Hour)).
		Order("timestamp ASC").
		Find(&rates).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get prices 24h: %w", err)
	}

	points := make([]model.PricePoint, 0, len(rates))
	for _, rate := range rates {
		points = append(points, model.PricePoint{
			Price:     model.CentsToFloat(rate.PriceUSDCents),
			Timestamp: rate.Timestamp,
		})
	}
	return points, nil
}
