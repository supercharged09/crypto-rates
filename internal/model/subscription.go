package model

import "time"

// Subscription — подписка пользователя Telegram на авто-рассылку
type Subscription struct {
	ID             int64     `json:"id"`
	ChatID         int64     `json:"chat_id"`
	Cryptocurrency string    `json:"cryptocurrency"`
	IntervalMin    int       `json:"interval_min"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
