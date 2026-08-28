package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/supercharged09/crypto-rates/internal/repository"
)

// StatsExporter — экспортирует статистику в JSON файл
type StatsExporter struct {
	userRepo *repository.UserRepository
	filePath string
}

// NewStatsExporter создаёт новый экспортер
func NewStatsExporter(userRepo *repository.UserRepository, filePath string) *StatsExporter {
	return &StatsExporter{
		userRepo: userRepo,
		filePath: filePath,
	}
}

// UserExport — структура для красивого JSON экспорта
type UserExport struct {
	ID           int64     `json:"id"`
	ChatID       int64     `json:"chat_id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	LanguageCode string    `json:"language_code"`
	FirstSeen    time.Time `json:"first_seen"`
	LastActive   time.Time `json:"last_active"`
	CommandCount int       `json:"command_count"`
}

// ExportUsers сохраняет всех пользователей в JSON файл
func (e *StatsExporter) ExportUsers() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users, err := e.userRepo.GetAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	// Преобразуем sql.NullString в обычные строки для JSON
	var exportUsers []UserExport
	for _, u := range users {
		exportUsers = append(exportUsers, UserExport{
			ID:           u.ID,
			ChatID:       u.ChatID,
			Username:     nullStringToString(u.Username),
			FirstName:    nullStringToString(u.FirstName),
			LastName:     nullStringToString(u.LastName),
			LanguageCode: nullStringToString(u.LanguageCode),
			FirstSeen:    u.FirstSeen,
			LastActive:   u.LastActive,
			CommandCount: u.CommandCount,
		})
	}

	export := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"total_users": len(exportUsers),
		"users":       exportUsers,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	if err := os.WriteFile(e.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("Users exported to %s (%d users)", e.filePath, len(exportUsers))
	return nil
}

// StartPeriodicExport запускает периодический экспорт
func (e *StatsExporter) StartPeriodicExport(ctx context.Context, interval time.Duration) {
	log.Printf("Stats exporter: exporting to %s every %v", e.filePath, interval)

	if err := e.ExportUsers(); err != nil {
		log.Printf("ERROR: export failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := e.ExportUsers(); err != nil {
				log.Printf("ERROR: export failed: %v", err)
			}
		case <-ctx.Done():
			log.Println("Stats exporter stopped")
			return
		}
	}
}

// nullStringToString конвертирует sql.NullString в обычную строку
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
