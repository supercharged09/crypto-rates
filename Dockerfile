# Этап сборки
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /crypto-rates ./cmd/app

# Финальный образ
FROM alpine:3.19

WORKDIR /app

# Копируем бинарник
COPY --from=builder /crypto-rates .

# Копируем миграции
COPY migrations ./migrations

# Создаём папку для логов
RUN mkdir -p /app/logs

EXPOSE 8080

CMD ["./crypto-rates"]