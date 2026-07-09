# ---- Build stage ----
FROM golang:1.21-alpine AS builder

WORKDIR /src

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static, stripped binary for a tiny final image.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/nearfriend-bot .

# ---- Runtime stage ----
FROM alpine:3.20

# ca-certificates: HTTPS calls to the Telegram + translation APIs.
# tzdata: the bot resolves user timezones (e.g. Asia/Tehran) for wake-hours
#         and time-based achievements — without it LoadLocation fails.
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -u 10001 app

WORKDIR /app
COPY --from=builder /out/nearfriend-bot .

USER app

ENTRYPOINT ["/app/nearfriend-bot"]
