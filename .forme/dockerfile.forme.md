# Dockerfile (pattern)

> pattern for Dockerfile (golang)

---

```Dockerfile
# 1. Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/api

# 2. Minimal run stage
FROM alpine:3.18

WORKDIR /app

# Создаем пользователя
RUN addgroup -S app && adduser -S app -G app

# Копируем бинарь из builder
COPY --from=builder /app/server .

USER app

EXPOSE 8080

CMD ["./server"]

```