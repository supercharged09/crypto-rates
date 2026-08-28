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
	client   client.CoinGeckoAPI
	rateRepo repository.RateRepositoryInterface
}

// NewRateService создает новый сервис
func NewRateService(
	client client.CoinGeckoAPI,
	rateRepo repository.RateRepositoryInterface,
) *RateService {
	return &RateService{
		client:   client,
		rateRepo: rateRepo,
	}
}

//FetchAndSaveRates получает курсы из API и сохраняет в БД

func (s *RateService) FetchAndSaveRates(ctx context.Context) error {
	log.Println("Fetching rates from CoinGecko...")

	prices, err := s.client.GetPrices()
	if err != nil {
		return fmt.Errorf("failed to fetch prices: %w", err)
	}

	now := time.Now()

	for _, crypto := range model.SupportedCryptos {
		price, ok := prices[crypto.ID]
		if !ok {
			log.Printf("WARNING: no price for %s in response", crypto.ID)
			continue
		}

		rate := model.Rate{
			Cryptocurrency: crypto.ID,
			PriceUSD:       price,
			Timestamp:      now,
		}
		if err := s.rateRepo.Save(ctx, rate); err != nil {
			log.Printf("ERROR: failed to save %s rate: %v", crypto.ID, err)
		}
	}
	log.Printf("Rates saved for %d cryptocurrencies", len(prices))
	return nil
}

// GetRateStats получает полную статистику по криптовалюте
func (s *RateService) GetRateStats(ctx context.Context, cryptocurrency string) (*model.RateStats, error) {
	//текущий курс
	current, err := s.rateRepo.GetCurrentPrice(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, fmt.Errorf("no data for %s", cryptocurrency)
	}

	//Min/Max за 24 часа
	min, max, err := s.rateRepo.GetMinMax24h(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}

	//изменение за час
	priceHourAgo, err := s.rateRepo.GetPriceHourAgo(ctx, cryptocurrency)
	if err != nil {
		return nil, err
	}

	var changePercent float64
	if priceHourAgo > 0 {
		changePercent = ((current.PriceUSD - priceHourAgo) / priceHourAgo) * 100
	}

	return &model.RateStats{
		Cryptocurrency:  cryptocurrency,
		CurrentPrice:    current.PriceUSD,
		MinPrice24h:     min,
		MaxPrice24h:     max,
		ChangePercent1h: changePercent,
		LastUpdated:     current.Timestamp,
	}, nil
}

// StartBackgroundUpdater запускает фоновое обновление курсов
func (s *RateService) StartBackgroundUpdater(ctx context.Context, interval time.Duration) {
	//первый запуск сразу
	if err := s.FetchAndSaveRates(ctx); err != nil {
		log.Printf("ERROR: initial fetch failed: %v", err)
	}

	//Тикер для периодического обновления
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
		MinPrice24h:     price, // Нет исторических данных
		MaxPrice24h:     price,
		ChangePercent1h: 0,
		LastUpdated:     time.Now(),
	}, nil
}
