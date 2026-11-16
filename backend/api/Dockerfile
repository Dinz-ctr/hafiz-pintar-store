
# =========================================
# BUILD STAGE
# =========================================
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Install Goose migration CLI
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.30.1

# Download Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build API
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/hafiz-api ./cmd/server


# =========================================
# RUNTIME STAGE
# =========================================
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# Copy API binary
COPY --from=builder /app/hafiz-api ./hafiz-api

# Copy Goose CLI
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy database migrations
COPY db/migrations/ ./db/migrations/

EXPOSE 8080

CMD ["./hafiz-api"]
