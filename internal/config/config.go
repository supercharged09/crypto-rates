package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config — корневая структура конфигурации
type Config struct {
	Database  DatabaseConfig
	CoinGecko CoinGeckoConfig
	Telegram  TelegramConfig
	Server    ServerConfig
	Logging   LoggingConfig
	Service   ServiceConfig
}

// DatabaseConfig — настройки PostgreSQL
type DatabaseConfig struct {
	Host            string `env:"DB_HOST" env-required:"true"`
	Port            int    `env:"DB_PORT" env-required:"true"`
	User            string `env:"DB_USER" env-required:"true"`
	Password        string `env:"DB_PASSWORD" env-required:"true"`
	Name            string `env:"DB_NAME" env-required:"true"`
	SSLMode         string `env:"DB_SSLMODE" env-default:"disable"`
	MaxOpenConns    int    `env:"DB_MAX_OPEN_CONNS" env-default:"25"`
	MaxIdleConns    int    `env:"DB_MAX_IDLE_CONNS" env-default:"5"`
	ConnMaxLifetime string `env:"DB_CONN_MAX_LIFETIME" env-default:"5m"`
}

// DSN возвращает строку подключения для GORM
func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// CoinGeckoConfig — настройки CoinGecko API
type CoinGeckoConfig struct {
	BaseURL       string `env:"COINGECKO_URL" env-required:"true"`
	TimeoutSecond int    `env:"COINGECKO_TIMEOUT_SEC" env-required:"true"`
}

// TelegramConfig — настройки Telegram бота
type TelegramConfig struct {
	Token          string `env:"TELEGRAM_TOKEN" env-required:"true"`
	HTTPTimeoutSec int    `env:"TELEGRAM_HTTP_TIMEOUT_SEC" env-required:"true"`
	TLSTimeoutSec  int    `env:"TELEGRAM_TLS_TIMEOUT_SEC" env-required:"true"`
	DialTimeoutSec int    `env:"TELEGRAM_DIAL_TIMEOUT_SEC" env-required:"true"`
	UpdatesTimeout int    `env:"TELEGRAM_UPDATES_TIMEOUT" env-required:"true"`
}

// ServerConfig — настройки HTTP сервера
type ServerConfig struct {
	Port               string `env:"HTTP_SERVER_PORT" env-required:"true"`
	ReadTimeoutSec     int    `env:"HTTP_READ_TIMEOUT_SEC" env-required:"true"`
	WriteTimeoutSec    int    `env:"HTTP_WRITE_TIMEOUT_SEC" env-required:"true"`
	IdleTimeoutSec     int    `env:"HTTP_IDLE_TIMEOUT_SEC" env-required:"true"`
	ShutdownTimeoutSec int    `env:"HTTP_SHUTDOWN_TIMEOUT_SEC" env-required:"true"`
}

// LoggingConfig — настройки логирования
type LoggingConfig struct {
	File         string `env:"LOG_FILE" env-required:"true"`
	MaxSizeMB    int    `env:"LOG_MAX_SIZE" env-required:"true"`
	MaxBackups   int    `env:"LOG_MAX_BACKUPS" env-required:"true"`
	MaxAgeDays   int    `env:"LOG_MAX_AGE" env-required:"true"`
	Compress     bool   `env:"LOG_COMPRESS" env-required:"true"`
	AlsoToStdout bool   `env:"LOG_TO_STDOUT" env-required:"true"`
}

// ServiceConfig — настройки самого сервиса (интервалы фоновых задач)
type ServiceConfig struct {
	RatesUpdateInterval    string `env:"RATES_UPDATE_INTERVAL" env-required:"true"`
	SchedulerIntervalMin   int    `env:"SCHEDULER_INTERVAL_MIN" env-required:"true"`
	AlertCheckIntervalSec  int    `env:"ALERT_CHECK_INTERVAL_SEC" env-required:"true"`
	StatsExportIntervalMin int    `env:"STATS_EXPORT_INTERVAL_MIN" env-required:"true"`
	StatsExportPath        string `env:"STATS_EXPORT_PATH" env-required:"true"`
}

// MustLoad загружает конфиг или падает
func MustLoad() *Config {
	var cfg Config

	envPath := findEnvFile()
	if envPath != "" {
		// .env найден — читаем из него
		if err := cleanenv.ReadConfig(envPath, &cfg); err != nil {
			panic(fmt.Sprintf("config error reading %s: %s", envPath, err))
		}
		log.Printf("Loaded config from: %s", envPath)
	} else {
		// .env не найден — читаем из переменных окружения (для Docker/Railway)
		log.Println("Config: .env not found, using environment variables")
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			panic(fmt.Sprintf("config error from env: %s", err))
		}
	}

	// читаем ServiceConfig отдельно
	var svc ServiceConfig
	if envPath != "" {
		if err := cleanenv.ReadConfig(envPath, &svc); err != nil {
			panic(fmt.Sprintf("config error reading service config: %s", err))
		}
	} else {
		if err := cleanenv.ReadEnv(&svc); err != nil {
			panic(fmt.Sprintf("config error reading service env: %s", err))
		}
	}
	cfg.Service = svc

	return &cfg
}

// findProjectRoot поднимается до директории с go.mod
func findProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// fileExists проверяет существование файла
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
