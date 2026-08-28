# Этап сборки
FROM golang:1.25-alpine AS builder
# или же FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /crypto-rates ./cmd/app

# Устанавливаем migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Финальный образ
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /crypto-rates .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY migrations ./migrations

RUN mkdir -p /app/logs

EXPOSE 8080

# Скрипт запуска: сначала миграции, потом приложение
CMD ["sh", "-c", "migrate -path migrations -database \"$DATABASE_URL\" up && ./crypto-rates"]

