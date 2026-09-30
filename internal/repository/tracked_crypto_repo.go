package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/supercharged09/crypto-rates/internal/model"
)

// TrackedCryptoRepository — работа с таблицей tracked_cryptos
type TrackedCryptoRepository struct {
	db *gorm.DB
}

// NewTrackedCryptoRepository создаёт новый репозиторий
func NewTrackedCryptoRepository(db *gorm.DB) *TrackedCryptoRepository {
	return &TrackedCryptoRepository{db: db}
}

// Add добавляет монету в отслеживаемые (если ещё не добавлена)
func (r *TrackedCryptoRepository) Add(ctx context.Context, crypto model.TrackedCrypto) error {
	err := r.db.WithContext(ctx).
		Where("coin_id = ?", crypto.CoinID).
		FirstOrCreate(&crypto).Error

	if err != nil {
		return fmt.Errorf("failed to add tracked crypto: %w", err)
	}
	return nil
}

// GetAll возвращает все отслеживаемые монеты
func (r *TrackedCryptoRepository) GetAll(ctx context.Context) ([]model.TrackedCrypto, error) {
	var cryptos []model.TrackedCrypto
	err := r.db.WithContext(ctx).
		Order("created_at ASC").
		Find(&cryptos).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get tracked cryptos: %w", err)
	}
	return cryptos, nil
}

// Exists проверяет, отслеживается ли монета
func (r *TrackedCryptoRepository) Exists(ctx context.Context, coinID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.TrackedCrypto{}).
		Where("coin_id = ?", coinID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check tracked crypto: %w", err)
	}
	return count > 0, nil
}

// GetByCoinID возвращает монету по CoinGecko ID
func (r *TrackedCryptoRepository) GetByCoinID(ctx context.Context, coinID string) (*model.TrackedCrypto, error) {
	var crypto model.TrackedCrypto
	err := r.db.WithContext(ctx).
		Where("coin_id = ?", coinID).
		First(&crypto).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked crypto: %w", err)
	}
	return &crypto, nil
}

// Delete удаляет монету из отслеживаемых
func (r *TrackedCryptoRepository) Delete(ctx context.Context, coinID string) error {
	err := r.db.WithContext(ctx).
		Where("coin_id = ?", coinID).
		Delete(&model.TrackedCrypto{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete tracked crypto: %w", err)
	}
	return nil
}

// IsBaseCrypto проверяет, является ли монета базовой (из SupportedCryptos)
func IsBaseCrypto(coinID string) bool {
	for _, c := range model.SupportedCryptos {
		if c.ID == coinID {
			return true
		}
	}
	return false
}
