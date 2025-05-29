FROM golang:latest as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o go-cron-webhooks

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /app/go-cron-webhooks .

ARG DATABASE_URL
ENV DATABASE_URL=${DATABASE_URL}

ARG PORT=3000
ENV PORT=${PORT}

EXPOSE ${PORT}
CMD ["/app/go-cron-webhooks"]