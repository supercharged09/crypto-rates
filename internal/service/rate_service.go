package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/supercharged09/crypto-rates/internal/client"
	"github.com/supercharged09/crypto-rates/internal/model"
	"github.com/supercharged09/crypto-rates/internal/repository"
)

// RateServiceInterface интерфейс для теста
type RateServiceInterface interface {
	GetRateStats(ctx context.Context, cryptocurrency string) (*model.RateStats, error)
	GetAnyRate(ctx context.Context, cryptoID string) (*model.RateStats, error)
}

// RateService - бизнес логика работы с курсами
type RateService struct {
	client      client.CoinGeckoAPI
	rateRepo    repository.RateRepositoryInterface
	trackedRepo *repository.TrackedCryptoRepository
}

// NewRateService создает новый сервис
func NewRateService(
	client client.CoinGeckoAPI,
	rateRepo repository.RateRepositoryInterface,
	trackedRepo *repository.TrackedCryptoRepository,
) *RateService {
	return &RateService{
		client:      client,
		rateRepo:    rateRepo,
		trackedRepo: trackedRepo,
	}
}

// FetchAndSaveRates получает курсы из API и сохраняет в БД
func (s *RateService) FetchAndSaveRates(ctx context.Context) error {
	log.Println("Fetching rates from CoinGecko...")

	// Собираем список ID: базовые + произвольные
	ids := make([]string, 0, len(model.SupportedCryptos))
	for _, crypto := range model.SupportedCryptos {
		ids = append(ids, crypto.ID)
	}

	// Если есть репозиторий произвольных монет — добавляем их
	if s.trackedRepo != nil {
		tracked, err := s.trackedRepo.GetAll(ctx)
		if err != nil {
			log.Printf("WARNING: failed to get tracked cryptos: %v", err)
		} else {
			for _, tc := range tracked {
				ids = append(ids, tc.CoinID)
			}
		}
	}

	// Один запрос ко всем монетам
	prices, err := s.client.GetPricesForIDs(ids)
	if err != nil {
		return fmt.Errorf("failed to fetch prices: %w", err)
	}

	now := time.Now()
	saved := 0

	for _, cryptoID := range ids {
		price, ok := prices[cryptoID]
		if !ok {
			log.Printf("WARNING: no price for %s in response", cryptoID)
			continue
		}

		rate := model.Rate{
			Cryptocurrency: cryptoID,
			PriceUSDCents:  model.FloatToCents(price),
			Timestamp:      now,
		}
		if err := s.rateRepo.Save(ctx, rate); err != nil {
			log.Printf("ERROR: failed to save %s rate: %v", cryptoID, err)
			continue
		}
		saved++
	}

	log.Printf("Rates saved for %d cryptocurrencies", saved)
	return nil
}

// GetRateStats получает полную статистику по криптовалюте
func (s *RateService) GetRateStats(ctx context.Context, cryptocurrency string) (*model.RateStats, error) {
	// текущий курс
	current, err := s.rateRepo.GetCurrentPrice(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("no data for %s", cryptocurrency)
	}

	// конвертируем копейки в доллары
	currentPrice := model.CentsToFloat(current.PriceUSDCents)

	// Min/Max за 24 часа
	min, max, err := s.rateRepo.GetMinMax24h(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}

	// изменение за час
	priceHourAgo, err := s.rateRepo.GetPriceHourAgo(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}

	var changePercent float64
	if priceHourAgo > 0 {
		changePercent = ((currentPrice - priceHourAgo) / priceHourAgo) * 100
	}

	return &model.RateStats{
		Cryptocurrency:  cryptocurrency,
		CurrentPrice:    currentPrice,
		MinPrice24h:     min,
		MaxPrice24h:     max,
		ChangePercent1h: changePercent,
		LastUpdated:     current.Timestamp,
	}, nil
}

// StartBackgroundUpdater запускает фоновое обновление курсов
func (s *RateService) StartBackgroundUpdater(ctx context.Context, interval time.Duration) {
	// первый запуск сразу
	if err := s.FetchAndSaveRates(ctx); err != nil {
		log.Printf("ERROR: initial fetch failed: %v", err)
	}

	// Тикер для периодического обновления
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.FetchAndSaveRates(ctx); err != nil {
				log.Printf("ERROR: fetch failed: %v", err)
			}
		case <-ctx.Done():
			log.Println("Background updater stopped")
			return
		}
	}
}

// GetAnyRate получает текущую цену произвольной монеты с CoinGecko
func (s *RateService) GetAnyRate(ctx context.Context, cryptoID string) (*model.RateStats, error) {
	// Пробуем сначала из БД (если уже сохраняли)
	stats, err := s.GetRateStats(ctx, cryptoID)
	if err == nil {
		return stats, nil
	}

	// Если нет в БД — запрашиваем напрямую с CoinGecko
	price, err := s.client.GetAnyPrice(cryptoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get price for %s: %w", cryptoID, err)
	}

	return &model.RateStats{
		Cryptocurrency:  cryptoID,
		CurrentPrice:    price,
		MinPrice24h:     price,
		MaxPrice24h:     price,
		ChangePercent1h: 0,
		LastUpdated:     time.Now(),
	}, nil
}
