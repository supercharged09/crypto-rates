package model

import (
	"database/sql"
	"time"
)

// User — информация о пользователе бота
type User struct {
	ID           int64          `json:"id"`
	ChatID       int64          `json:"chat_id"`
	Username     sql.NullString `json:"username"`
	FirstName    sql.NullString `json:"first_name"`
	LastName     sql.NullString `json:"last_name"`
	LanguageCode sql.NullString `json:"language_code"`
	FirstSeen    time.Time      `json:"first_seen"`    //дата первого обращения
	LastActive   time.Time      `json:"last_active"`   //дата последней активности
	CommandCount int            `json:"command_count"` //общее кол-во отправленных команд
}
