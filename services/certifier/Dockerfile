# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy workspace and module files
COPY go.work go.mod go.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY shared/ ./shared/
COPY services/calculator/go.mod services/calculator/go.sum ./services/calculator/
COPY services/certifier/go.mod services/certifier/go.sum ./services/certifier/
COPY services/reporting/go.mod services/reporting/go.sum ./services/reporting/
COPY services/tracker/go.mod services/tracker/go.sum ./services/tracker/
COPY services/user-auth/go.mod services/user-auth/go.sum ./services/user-auth/
COPY services/wallet/go.mod services/wallet/go.sum ./services/wallet/

# Download dependencies with workspace support
ENV GOWORK=/app/go.work
RUN cd services/certifier && go mod download

# Copy source code
COPY services/certifier/ ./services/certifier/

# Build the application
RUN cd services/certifier && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/bin/certifier \
    ./cmd/main.go

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/certifier /app/certifier

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8086 9086

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8086/health || exit 1

# Run the application
CMD ["./certifier"]
