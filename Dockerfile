FROM golang:latest as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o go-cron-webhooks

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /app/go-cron-webhooks .
EXPOSE 3000
CMD ["/app/go-cron-webhooks"]