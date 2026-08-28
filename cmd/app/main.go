package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/supercharged09/crypto-rates/internal/logger"

	"github.com/supercharged09/crypto-rates/internal/bot"
	"github.com/supercharged09/crypto-rates/internal/client"
	"github.com/supercharged09/crypto-rates/internal/config"
	"github.com/supercharged09/crypto-rates/internal/handler"
	"github.com/supercharged09/crypto-rates/internal/repository"
	"github.com/supercharged09/crypto-rates/internal/service"
)

func main() {
	cfg := config.MustLoad()
	//настройка логгирования
	logger.Setup(cfg.ToLoggerConfig())
	log.Println("Config loaded successfully")

	// Подключаемся к БД
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// Создаём компоненты
	coinGeckoClient := client.NewCoinGeckoClient(cfg.ExternalAPIURL)
	rateRepo := repository.NewRateRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)
	userRepo := repository.NewUserRepository(db)
	rateService := service.NewRateService(coinGeckoClient, rateRepo)
	analyticsService := service.NewAnalyticsService(userRepo)
	chartService := service.NewChartService(rateRepo)
	rateHandler := handler.NewRateHandler(rateService, analyticsService, chartService)

	//HTTP роутер
	router := handler.NewRouter(rateHandler)

	//telegram bot
	cryptoBot, err := bot.NewCryptoBot(cfg.TelegramToken, rateService, subRepo, analyticsService, chartService)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	//планировщик авто рассылки
	scheduler := service.NewScheduler(subRepo, cryptoBot)

	// Экспортёр статистики пользователей
	statsExporter := service.NewStatsExporter(userRepo, "logs/users.json")

	// контекст для фоновых горутин
	ctxBg, cancelBg := context.WithCancel(context.Background())
	defer cancelBg()

	//фоновое обновление курсов
	go rateService.StartBackgroundUpdater(ctxBg, cfg.ParseUpdateInterval())
	go statsExporter.StartPeriodicExport(ctxBg, 5*time.Minute)

	//HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPServerPort),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запускаем сервер в горутине и сразу проверяем ошибку
	go func() {
		log.Printf("HTTP server starting on port %s", cfg.HTTPServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Используем Printf вместо Fatalf — не роняем всё приложение
			log.Printf("HTTP server error: %v", err)
		}
	}()
	// Даём серверу 100 мс на старт и проверяем
	time.Sleep(100 * time.Millisecond)
	log.Printf("HTTP server should be listening on :%s", cfg.HTTPServerPort)

	//Запуск бота
	go cryptoBot.Start(ctxBg)
	//запуск планировщика рассылки
	go scheduler.Start(ctxBg)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Service started. Press Ctrl+C to stop.")
	<-quit

	log.Println("Shutting down...")

	//Даем серверу 10 секунд на завершение текущих запросов
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	//Останавливаем фоновое обновление
	cancelBg()
	log.Println("Service stopped")
}
