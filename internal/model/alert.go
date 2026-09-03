package model

import "time"

// Alert — уведомление о пороге цены
type Alert struct {
	ID             int64      `json:"id"`
	ChatID         int64      `json:"chat_id"`
	Cryptocurrency string     `json:"cryptocurrency"`
	Direction      string     `json:"direction"` // "above" или "below"
	PriceThreshold float64    `json:"price_threshold"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	TriggeredAt    *time.Time `json:"triggered_at,omitempty"`
}
