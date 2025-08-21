FROM golang:1.25-alpine AS builder

# Build arguments
ARG VERSION=dev
ARG BUILD_TIME

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o iletken

# Final stage
FROM cgr.dev/chainguard/wolfi-base:latest

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates wget

# Create non-root user
RUN addgroup -g 1001 -S iletken && \
    adduser -u 1001 -S iletken -G iletken

# Set working directory
WORKDIR /app

# Copy binary and config
COPY --from=builder /app/iletken .
COPY --from=builder /app/iletken.yml .
COPY --from=builder /app/index.html .

# Change ownership
RUN chown -R iletken:iletken /app

# Switch to non-root user
USER iletken

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./iletken"]
