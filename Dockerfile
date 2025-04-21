FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/avito-service ./cmd/app

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime && \
    echo "Europe/Moscow" > /etc/timezone && \
    apk del tzdata

WORKDIR /app

COPY --from=builder /app/avito-service .
COPY --from=builder /app/migrations ./migrations

ENV HTTP_ADDR=:8080 \
    GRPC_ADDR=:3000 \
    PROMETHEUS_ADDR=:9000 \
    DB_URL="postgres://postgres:postgres@db:5432/avito?sslmode=disable" \
    JWT_SECRET="supersecretkey"

EXPOSE 8080 3000 9000

CMD ["/app/avito-service"] 