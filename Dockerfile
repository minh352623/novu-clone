# Stage 1: Builder
FROM golang:1.24-alpine3.20 AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Install Goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy go.mod and go.sum first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application with optimized flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tek-notification ./cmd/drunk

# Stage 2: Runtime
FROM alpine:3.20 

# Install necessary utilities
RUN apk add --no-cache bash

# Create non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create storages directory with correct permissions
RUN mkdir -p /storages/logs && chown appuser:appgroup /storages/logs

# Copy configuration and schema files
COPY --chown=appuser:appgroup ./config /config
COPY --chown=appuser:appgroup ./sql/schema /sql/schema

# Copy the built binary and Goose from builder stage
COPY --from=builder --chown=appuser:appgroup /build/tek-notification /
COPY --from=builder --chown=appuser:appgroup /go/bin/goose /usr/local/bin/goose
COPY --chown=appuser:appgroup entrypoint.sh /entrypoint.sh

# Set executable permissions
RUN chmod +x /entrypoint.sh

# Run as non-root user
USER appuser

ENTRYPOINT ["/entrypoint.sh"]
