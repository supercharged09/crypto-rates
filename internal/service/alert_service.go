package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/supercharged09/crypto-rates/internal/model"
	"github.com/supercharged09/crypto-rates/internal/repository"
)

// AlertService сервис алертов
type AlertService struct {
	alertRepo *repository.AlertRepository
}

// NewAlertService новый сервис алертов
func NewAlertService(alertRepo *repository.AlertRepository) *AlertService {
	return &AlertService{alertRepo: alertRepo}
}

// CreateAlert создает новый алерт
func (s *AlertService) CreateAlert(ctx context.Context, chatID int64, cryptocurrency, direction string, threshold float64) error {
	if direction != "above" && direction != "below" {
		return fmt.Errorf("direction must be 'above' or 'below'")
	}
	if threshold <= 0 {
		return fmt.Errorf("threshold must be positive")
	}

	alert := model.Alert{
		ChatID:              chatID,
		Cryptocurrency:      cryptocurrency,
		Direction:           direction,
		PriceThresholdCents: model.FloatToCents(threshold),
		IsActive:            true,
	}

	return s.alertRepo.Create(ctx, alert)
}

// GetUserAlerts возвращает все алерты пользователя
func (s *AlertService) GetUserAlerts(ctx context.Context, chatID int64) ([]model.Alert, error) {
	return s.alertRepo.GetByChatID(ctx, chatID)
}

// DeactivateAlert отключает алерт
func (s *AlertService) DeactivateAlert(ctx context.Context, id int64) error {
	return s.alertRepo.Deactivate(ctx, id)
}

// AlertSender нтерфейс для отправки уведомлений
type AlertSender interface {
	SendAlert(chatID int64, message string)
}

// StartAlertChecker запускает периодическую проверку
func (s *AlertService) StartAlertChecker(ctx context.Context, rateSvc *RateService, sender AlertSender) {
	log.Println("Alert checker started: checking every 30 seconds")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAlerts(ctx, rateSvc, sender)
		case <-ctx.Done():
			log.Println("Alert checker stopped")
			return
		}
	}
}

// checkAlerts проверяет все активные алерты
func (s *AlertService) checkAlerts(ctx context.Context, rateSvc *RateService, sender AlertSender) {
	alerts, err := s.alertRepo.GetActive(ctx)
	if err != nil {
		log.Printf("ERROR: failed to get alerts: %v", err)
		return
	}

	for _, alert := range alerts {
		stats, err := rateSvc.GetRateStats(ctx, alert.Cryptocurrency)
		if err != nil {
			continue
		}

		//порог из центов в  дуллеры
		threshold := model.CentsToFloat(alert.PriceThresholdCents)

		triggered := false
		if alert.Direction == "above" && stats.CurrentPrice >= threshold {
			triggered = true
		}
		if alert.Direction == "below" && stats.CurrentPrice <= threshold {
			triggered = true
		}

		if triggered {
			message := fmt.Sprintf(
				"🔔 *Алерт сработал!*\n%s цена $%.2f %s порога $%.2f",
				alert.Cryptocurrency,
				stats.CurrentPrice,
				map[string]string{"above": "выше", "below": "ниже"}[alert.Direction],
				threshold,
			)
			sender.SendAlert(alert.ChatID, message)

			if err := s.alertRepo.MarkTriggered(ctx, alert.ID); err != nil {
				log.Printf("ERROR: failed to mark alert %d: %v", alert.ID, err)
			}
		}
	}
}
