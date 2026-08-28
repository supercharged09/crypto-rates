package service

import (
	"context"
	"database/sql"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/supercharged09/crypto-rates/internal/model"
	"github.com/supercharged09/crypto-rates/internal/repository"
)

// AnalyticsService сервис аналитики пользователей
type AnalyticsService struct {
	userRepo *repository.UserRepository
}

// NewAnalyticsService создает новый сервис аналитики
func NewAnalyticsService(userRepo *repository.UserRepository) *AnalyticsService {
	return &AnalyticsService{userRepo: userRepo}
}

// TrackActivity отслеживание активности пользователя (обычные сообщения)
func (s *AnalyticsService) TrackActivity(msg *tgbotapi.Message) {
	if msg == nil || msg.From == nil {
		return
	}

	user := model.User{
		ChatID: msg.Chat.ID,
		Username: sql.NullString{
			String: msg.From.UserName,
			Valid:  msg.From.UserName != "",
		},
		FirstName: sql.NullString{
			String: msg.From.FirstName,
			Valid:  msg.From.FirstName != "",
		},
		LastName: sql.NullString{
			String: msg.From.LastName,
			Valid:  msg.From.LastName != "",
		},
		LanguageCode: sql.NullString{
			String: msg.From.LanguageCode,
			Valid:  msg.From.LanguageCode != "",
		},
	}

	if err := s.userRepo.Upsert(context.Background(), user); err != nil {
		log.Printf("ERROR: failed to track user activity: %v", err)
	}
}

// TrackCallbackActivity отслеживает активность из callback (inline-кнопки)
func (s *AnalyticsService) TrackCallbackActivity(callback *tgbotapi.CallbackQuery) {
	if callback == nil || callback.From == nil {
		return
	}

	// Используем callback.From — это РЕАЛЬНЫЙ пользователь
	// callback.Message.From — это БОТ (неправильно!)
	user := model.User{
		ChatID: callback.Message.Chat.ID,
		Username: sql.NullString{
			String: callback.From.UserName,
			Valid:  callback.From.UserName != "",
		},
		FirstName: sql.NullString{
			String: callback.From.FirstName,
			Valid:  callback.From.FirstName != "",
		},
		LastName: sql.NullString{
			String: callback.From.LastName,
			Valid:  callback.From.LastName != "",
		},
		LanguageCode: sql.NullString{
			String: callback.From.LanguageCode,
			Valid:  callback.From.LanguageCode != "",
		},
	}

	if err := s.userRepo.Upsert(context.Background(), user); err != nil {
		log.Printf("ERROR: failed to track callback activity: %v", err)
	}
}

// GetStats возвращает статистику
func (s *AnalyticsService) GetStats(ctx context.Context) (*repository.UserStats, error) {
	return s.userRepo.GetStats(ctx)
}
