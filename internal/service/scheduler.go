package service

import (
	"context"
	"log"
	"time"

	"github.com/supercharged09/crypto-rates/internal/repository"
)

// RateSender интерфейс для отправки курсов. Реализует bot.CryptoBot
type RateSender interface {
	SendRateToChat(chatID int64)
}

// Scheduler планировщик авто рассылки
type Scheduler struct {
	subRepo *repository.SubscriptionRepository
	sender  RateSender
}

// NewScheduler создает новый планировщие
func NewScheduler(subRepo *repository.SubscriptionRepository, sender RateSender) *Scheduler {
	return &Scheduler{
		subRepo: subRepo,
		sender:  sender,
	}
}

// Start запускает проверку подписок каждую минуту
func (s *Scheduler) Start(ctx context.Context) {
	log.Println("Scheduler started: checking subscriptions every minute")

	//Проверка каждую минуту
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processSubscriptions()

		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		}
	}
}

// processSubscriptions находит подписки, готовые к отправке, и отправляет курсы
func (s *Scheduler) processSubscriptions() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//находим подписки которые пора отправить
	subs, err := s.subRepo.GetDueSubscriptions(ctx)
	if err != nil {
		log.Printf("ERROR: sceduler failed to get subscriptions: %v", err)
		return
	}

	if len(subs) == 0 {
		return //не ждет рассылку никто
	}

	log.Printf("Scheduler: sending rates to %d subscribers", len(subs))

	//группировка по чат айди, чтобы не отправлять одному чату дважды
	sent := make(map[int64]bool)

	for _, sub := range subs {
		//проверка, не отправлено ли уже этому чату
		if sent[sub.ChatID] {
			//обновление updated_at, чтобы не дергать повторно
			if err := s.subRepo.Touch(ctx, sub.ID); err != nil {
				log.Printf("ERROR: failed to touch subscription %d: %v", sub.ID, err)
			}
			continue
		}

		//отправка в интерфейс, не зная, что это бот
		s.sender.SendRateToChat(sub.ChatID)
		sent[sub.ChatID] = true

		if err := s.subRepo.Touch(ctx, sub.ID); err != nil {
			log.Printf("ERROR: failed to touch subscription %d: %v", sub.ID, err)
		}
	}
}
