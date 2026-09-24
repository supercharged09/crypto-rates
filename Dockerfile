# Этап сборки
FROM golang:1.25-alpine AS builder
# или же FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /crypto-rates ./cmd/app

# Финальный образ
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /crypto-rates .

RUN mkdir -p /app/logs

EXPOSE 8080

# Запуск приложения (миграции теперь выполняются через GORM AutoMigrate)
CMD ["./crypto-rates"]

