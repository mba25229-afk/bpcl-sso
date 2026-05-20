# ── Stage 1: Build the Go binary ──────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install migrate CLI
RUN apk add --no-cache git && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/api ./cmd/api

# ── Stage 2: Runtime image ───────────────────────────────────────────────────
FROM alpine:3.21

RUN apk --no-cache add ca-certificates bash postgresql-client

WORKDIR /app

# Copy binary and migrate CLI from builder
COPY --from=builder /app/bin/api /app/bin/api
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Copy migrations
COPY migrations /app/migrations

# Copy entrypoint script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/bin/api"]
