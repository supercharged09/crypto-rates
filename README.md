# Crypto Rates

Сервис для отслеживания курсов криптовалют с REST API и Telegram ботом.

## Возможности

- Получение курсов BTC, ETH, USDT, BNB, XRP, SOL с CoinGecko
- Авто-обновление курсов каждые 5 минут
- Отслеживание любых монет с CoinGecko по запросу
- Статистика: текущая цена, min/max за 24ч, изменение за час
- Графики цен за 24 часа (HTML)
- Telegram бот с командами и кнопками
- Авто-рассылка курсов по расписанию
- Алерты при достижении порога цены
- Аналитика пользователей

## Стек

- **Go 1.25** — язык
- **PostgreSQL 16** — база данных
- **GORM** — ORM с автомиграциями
- **chi** — HTTP роутер
- **go-telegram-bot-api** — Telegram бот
- **go-echarts** — графики
- **Docker Compose** — локальное окружение
- **Swagger** — документация API

## Быстрый старт

### Требования

- Docker
- Go 1.25+
- [migrate](https://github.com/golang-migrate/migrate) — опционально

### Запуск

1. Клонируй репозиторий:
   ```bash
   git clone https://github.com/supercharged09/crypto-rates.git
   cd crypto-rates
   
2. Создай .env в корне проекта:

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=crypto
DB_PASSWORD=crypto
DB_NAME=crypto_rates
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# CoinGecko
COINGECKO_URL=https://api.coingecko.com/api/v3
COINGECKO_API_KEY=твой_demo_ключ
COINGECKO_TIMEOUT_SEC=30

# Telegram
TELEGRAM_TOKEN=твой_токен_бота
TELEGRAM_HTTP_TIMEOUT_SEC=30
TELEGRAM_TLS_TIMEOUT_SEC=15
TELEGRAM_DIAL_TIMEOUT_SEC=10
TELEGRAM_UPDATES_TIMEOUT=60

# HTTP Server
HTTP_SERVER_PORT=8080
HTTP_READ_TIMEOUT_SEC=10
HTTP_WRITE_TIMEOUT_SEC=10
HTTP_IDLE_TIMEOUT_SEC=60
HTTP_SHUTDOWN_TIMEOUT_SEC=10

# Logging
LOG_FILE=logs/crypto-rates.log
LOG_MAX_SIZE=10
LOG_MAX_BACKUPS=5
LOG_MAX_AGE=30
LOG_COMPRESS=true
LOG_TO_STDOUT=true

# Service
RATES_UPDATE_INTERVAL=5m
SCHEDULER_INTERVAL_MIN=1
ALERT_CHECK_INTERVAL_SEC=30
STATS_EXPORT_INTERVAL_MIN=5
STATS_EXPORT_PATH=logs/users.json

3. Запусти PostgreSQL:
```bash
docker compose up -d
```
4. Запусти приложение:
```bash
go run cmd/app/main.go
```

Таблицы создаются автоматически через GORM.

REST API

Базовый URL: http://localhost:8080


/health - Проверка статуса
/rates - 	Все курсы
/rates/{crypto} - Курс конкретной монеты
/chart/{crypto} - HTML-график за 24ч
/analytics - Статистика пользователей

Документация: http://localhost:8080/swagger/index.html

Примеры:
```bash
curl http://localhost:8080/rates
curl http://localhost:8080/rates/bitcoin
curl http://localhost:8080/rates/dogecoin
```
Telegram бот

Команды:

/start	Приветствие и меню
/help	Список команд
/rates	Все курсы
/rates_btc	Курс BTC
/rates btc	То же самое
/rates shiba-inu	Для монет с дефисом
/tracked	Отслеживаемые монеты
/untrack dogecoin	Убрать монету
/alert btc above 70000	Алерт при цене выше порога
/alert eth below 1500	Алерт при цене ниже порога
/alerts	Список алертов
/delalert 5	Удалить алерт
/start_auto 10	Авто-рассылка каждые 10 мин
/stop_auto	Отключить рассылку
/stats	Статистика сервиса

Разработка

Тесты:
```bash
go test ./...
```
Сборка:
```bash
go build ./...
```
Генерация Swagger:
```bash
swag init -g cmd/app/main.go --parseDependency
```
