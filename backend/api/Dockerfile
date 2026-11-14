
# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy Go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/hafiz-api ./cmd/server

# Runtime stage
FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/hafiz-api ./hafiz-api

EXPOSE 8080

CMD ["./hafiz-api"]
