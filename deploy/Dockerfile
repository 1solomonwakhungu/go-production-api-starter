# =============================================================================
# Multi-stage Dockerfile for Go Production API
# Stage 1: Build the binary using golang:alpine
# Stage 2: Run in a minimal alpine image
# =============================================================================

# --- Build Stage ---
FROM golang:1.26.5-alpine AS builder

# Install git (needed for some Go module downloads) and ca-certificates
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go.mod and go.sum first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -o /build/bin/api \
    ./cmd/api/

# --- Runtime Stage ---
FROM alpine:3.22

# Install ca-certificates for HTTPS and tzdata for timezone support
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /build/bin/api .
# Copy migration files
COPY --from=builder /build/migrations ./migrations

# Switch to non-root user
USER appuser

EXPOSE 8080

# Health check using the /health endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/api"]
