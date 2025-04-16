# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем файлы модуля и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /avito-pvz-service ./cmd/server

# Финальный образ
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /avito-pvz-service .
EXPOSE 8080
CMD ["./avito-pvz-service"]
