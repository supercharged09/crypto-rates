package app

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

	"github.com/supercharged09/crypto-rates/internal/bot"
	"github.com/supercharged09/crypto-rates/internal/client"
	"github.com/supercharged09/crypto-rates/internal/config"
	"github.com/supercharged09/crypto-rates/internal/handler"
	"github.com/supercharged09/crypto-rates/internal/logger"
	"github.com/supercharged09/crypto-rates/internal/repository"
	"github.com/supercharged09/crypto-rates/internal/service"
)

// Start запускает приложение целиком
func Start(cfg *config.Config) {
	// настройка логирования
	logger.Setup(
		cfg.Logging.File,
		cfg.Logging.MaxSizeMB,
		cfg.Logging.MaxBackups,
		cfg.Logging.MaxAgeDays,
		cfg.Logging.Compress,
		cfg.Logging.AlsoToStdout,
	)
	log.Println("Config loaded successfully")

	// подключаемся к БД
	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)

	connMaxLifetime, err := time.ParseDuration(cfg.Database.ConnMaxLifetime)
	if err != nil {
		log.Fatalf("Invalid DB_CONN_MAX_LIFETIME: %v", err)
	}
	db.SetConnMaxLifetime(connMaxLifetime)

	// проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established")

	// Автомиграции GORM
	if err := db.AutoMigrate(
		&model.Rate{},
		&model.Subscription{},
		&model.User{},
		&model.Alert{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations applied")

	// создаём компоненты
	coinGeckoClient := client.NewCoinGeckoClient(cfg.CoinGecko)
	rateRepo := repository.NewRateRepository(db)
	subRepo := repository.NewSubscriptionRepository(db)
	userRepo := repository.NewUserRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	rateService := service.NewRateService(coinGeckoClient, rateRepo)
	analyticsService := service.NewAnalyticsService(userRepo)
	chartService := service.NewChartService(rateRepo)
	alertService := service.NewAlertService(alertRepo)
	rateHandler := handler.NewRateHandler(rateService, analyticsService, chartService)

	// HTTP роутер
	router := handler.NewRouter(rateHandler)

	// Telegram бот
	cryptoBot, err := bot.NewCryptoBot(cfg.Telegram, rateService, subRepo, analyticsService, chartService, alertService)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// планировщик авто-рассылки
	scheduler := service.NewScheduler(subRepo, cryptoBot)

	// экспортёр статистики пользователей
	statsExporter := service.NewStatsExporter(userRepo, cfg.Service.StatsExportPath)

	// контекст для фоновых горутин
	ctxBg, cancelBg := context.WithCancel(context.Background())
	defer cancelBg()

	// фоновые задачи
	ratesInterval, err := time.ParseDuration(cfg.Service.RatesUpdateInterval)
	if err != nil {
		log.Fatalf("Invalid RATES_UPDATE_INTERVAL: %v", err)
	}
	go rateService.StartBackgroundUpdater(ctxBg, ratesInterval)

	statsInterval := time.Duration(cfg.Service.StatsExportIntervalMin) * time.Minute
	go statsExporter.StartPeriodicExport(ctxBg, statsInterval)

	go cryptoBot.Start(ctxBg)
	go scheduler.Start(ctxBg)
	go alertService.StartAlertChecker(ctxBg, rateService, cryptoBot)

	// HTTP сервер
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSec) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSec) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSec) * time.Second,
	}

	go func() {
		log.Printf("HTTP server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Service started. Press Ctrl+C to stop.")
	<-quit

	log.Println("Shutting down...")
	cancelBg()

	shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeoutSec) * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	log.Println("Service stopped")
}
