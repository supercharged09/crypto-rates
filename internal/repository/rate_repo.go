package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// RateRepositoryInterface - интерфейс для тестирования
type RateRepositoryInterface interface {
	Save(ctx context.Context, rate model.Rate) error
	GetCurrentPrice(ctx context.Context, cryptocurrency string) (*model.Rate, error)
	GetMinMax24h(ctx context.Context, cryptocurrency string) (float64, float64, error)
	GetPriceHourAgo(ctx context.Context, cryptocurrency string) (float64, error)
	GetPrices24h(ctx context.Context, cryptocurrency string) ([]model.PricePoint, error)
}

// RateRepository - работа с таблицей rates
type RateRepository struct {
	db *sql.DB
}

// NewRateRepository создает новый репозиторий
func NewRateRepository(db *sql.DB) *RateRepository {
	return &RateRepository{db: db}
}

// Save сохраняет курс в БД
func (r *RateRepository) Save(ctx context.Context, rate model.Rate) error {
	query := `
INSERT INTO rates (cryptocurrency, price_usd, timestamp)
VALUES ($1, $2, $3)
`

	_, err := r.db.ExecContext(ctx, query,
		rate.Cryptocurrency,
		rate.PriceUSD,
		rate.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to save rate: %w", err)
	}

	return nil
}

// GetCurrentPrice получает последнюю цену для крипты
func (r *RateRepository) GetCurrentPrice(ctx context.Context, cryptocurrency string) (*model.Rate, error) {
	query := `
SELECT id, cryptocurrency, price_usd, timestamp
FROM rates
WHERE cryptocurrency = $1
ORDER BY timestamp DESC
LIMIT 1
`

	var rate model.Rate
	err := r.db.QueryRowContext(ctx, query, cryptocurrency).Scan(
		&rate.ID,
		&rate.Cryptocurrency,
		&rate.PriceUSD,
		&rate.Timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, nil //данных пока нет
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current price: %w", err)
	}

	return &rate, nil
}

// GetMixMax24h получает минимальный и максимальный курс за 24 часа
func (r *RateRepository) GetMinMax24h(ctx context.Context, cryptocurrency string) (float64, float64, error) {
	query := `
SELECT 
    COALESCE(MIN(price_usd), 0),
    COALESCE(MAX(price_usd), 0)
FROM rates
WHERE cryptocurrency = $1
AND timestamp >= NOW() - INTERVAL '24 hours'
`
	var min, max float64
	err := r.db.QueryRowContext(ctx, query, cryptocurrency).Scan(&min, &max)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get min/max: %w", err)
	}

	return min, max, nil
}

// GetPriceHourAgo получает цену час назад (для расчета изменения в %)
func (r *RateRepository) GetPriceHourAgo(ctx context.Context, cryptocurrency string) (float64, error) {
	query := `
SELECT price_usd
FROM rates
WHERE cryptocurrency = $1
AND timestamp <= NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC
LIMIT 1
`

	var price float64
	err := r.db.QueryRowContext(ctx, query, cryptocurrency).Scan(&price)
	if err == sql.ErrNoRows {
		return 0, nil //данных за час пока нет

	}
	if err != nil {
		return 0, fmt.Errorf("failed to get price hour ago: %w", err)
	}

	return price, nil
}

// GetPrices24h возвращает цены за последние 24 часа
func (r *RateRepository) GetPrices24h(ctx context.Context, cryptocurrency string) ([]model.PricePoint, error) {
	query := `
	SELECT price_usd, timestamp
	FROM rates
	WHERE cryptocurrency = $1
	AND timestamp >= NOW() - INTERVAL '24 hours'
	ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, cryptocurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to get prices 24h: %w", err)
	}
	defer rows.Close()

	var points []model.PricePoint
	for rows.Next() {
		var p model.PricePoint
		if err := rows.Scan(&p.Price, &p.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan price point: %w", err)
		}
		points = append(points, p)
	}

	return points, rows.Err()
}
