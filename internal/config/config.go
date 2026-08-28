package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/supercharged09/crypto-rates/internal/logger"
)

type Config struct {
	DatabaseURL    string `env:"DATABASE_URL" env-required:"true"`
	ExternalAPIURL string `env:"EXTERNAL_API_URL" env-default:"https://api.coingecko.com/api/v3"`
	TelegramToken  string `env:"TELEGRAM_TOKEN" env-required:"true"`
	UpdateInterval string `env:"UPDATE_INTERVAL" env-default:"5m"`
	HTTPServerPort string `env:"HTTP_SERVER_PORT" env-default:"8080"`

	//логирование
	LogFile       string `env:"LOG_FILE" env-default:"logs/crypto-rates.log"`
	LogMaxSize    int    `env:"LOG_MAX_SIZE" env-default:"10"`
	LogMaxBackups int    `env:"LOG_MAX_BACKUPS" env-default:"5"`
	LogMaxAge     int    `env:"LOG_MAX_AGE" env-default:"30"`
	LogToStdout   bool   `env:"LOG_TO_STDOUT" env-default:"true"`
}

// ToLoggerConfig конвертирует Config в logger.Config
func (c *Config) ToLoggerConfig() logger.Config {
	return logger.Config{
		LogFile:      c.LogFile,
		MaxSize:      c.LogMaxSize,
		MaxBackups:   c.LogMaxBackups,
		MaxAge:       c.LogMaxAge,
		Compress:     true,
		AlsoToStdout: c.LogToStdout,
	}
}

func MustLoad() *Config {
	var cfg Config

	// Ищем .env файл в нескольких местах
	envPath := findEnvFile()

	if envPath != "" {
		err := cleanenv.ReadConfig(envPath, &cfg)
		if err != nil {
			panic(fmt.Sprintf("config error reading %s: %s", envPath, err))
		}
		fmt.Printf("Loaded config from: %s\n", envPath)
	} else {
		fmt.Println(".env file not found, using environment variables only")
	}

	// Проверяем обязательные переменные
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic(fmt.Sprintf("config error: %s", err))
	}

	return &cfg
}

// findEnvFile ищет .env начиная с текущей директории и поднимаясь выше
func findEnvFile() string {
	// 1. Проверяем текущую директорию
	if fileExists(".env") {
		return ".env"
	}

	// 2. Ищем корень проекта (где go.mod)
	root := findProjectRoot()
	if root != "" {
		envPath := filepath.Join(root, ".env")
		if fileExists(envPath) {
			return envPath
		}
	}

	return ""
}

// findProjectRoot ищет директорию с go.mod
func findProjectRoot() string {
	// Начинаем с директории, где лежит этот файл (config.go)
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Дошли до корня файловой системы
			break
		}
		dir = parent
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil

}

// ParseUpdateInterval парсит строку интервала в Duration
func (c *Config) ParseUpdateInterval() time.Duration {
	d, err := time.ParseDuration(c.UpdateInterval)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}
