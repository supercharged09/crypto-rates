package model

import "time"

// Alert — уведомление о пороге цены
type Alert struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID              int64      `gorm:"not null;index:idx_alerts_chat_id" json:"chat_id"`
	Cryptocurrency      string     `gorm:"type:varchar(20);not null" json:"cryptocurrency"`
	Direction           string     `gorm:"type:varchar(5);not null;check:direction IN ('above','below')" json:"direction"`
	PriceThresholdCents int64      `gorm:"not null" json:"price_threshold_cents"`
	IsActive            bool       `gorm:"not null;default:true;index:idx_alerts_active" json:"is_active"`
	CreatedAt           time.Time  `gorm:"not null;default:now()" json:"created_at"`
	TriggeredAt         *time.Time `json:"triggered_at,omitempty"`
}

func (Alert) TableName() string {
	return "alerts"
}
