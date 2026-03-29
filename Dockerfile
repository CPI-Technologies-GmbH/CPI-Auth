# CPI Auth Core - Multi-stage Docker build
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files first for better cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -o /cpi-auth ./main.go

# Production stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 cpi-auth && \
    adduser -D -u 1001 -G cpi-auth cpi-auth

WORKDIR /app

COPY --from=builder /cpi-auth .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config.yaml ./config.yaml

RUN chown -R cpi-auth:cpi-auth /app

USER cpi-auth

EXPOSE 5050

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:5050/health || exit 1

ENTRYPOINT ["./cpi-auth"]
