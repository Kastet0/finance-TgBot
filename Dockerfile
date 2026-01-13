FROM golang:1.25.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /finance-bot ./cmd/bot

FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /finance-bot /app/finance-bot
COPY --from=builder /app/migrations /app/migrations

RUN chmod +x /app/finance-bot

CMD ["/app/finance-bot"]