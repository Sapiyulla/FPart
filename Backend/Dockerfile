FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd

FROM alpine:3.23

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /app/server .

USER app

EXPOSE 7000-9999

CMD ["./server"]

