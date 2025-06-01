# Build stage
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy workspace and module files
COPY go.work go.mod go.sum ./
COPY shared/go.mod shared/go.sum ./shared/
COPY services/calculator/go.mod services/calculator/go.sum ./services/calculator/
COPY services/certifier/go.mod services/certifier/go.sum ./services/certifier/
COPY services/reporting/go.mod services/reporting/go.sum ./services/reporting/
COPY services/tracker/go.mod services/tracker/go.sum ./services/tracker/
COPY services/user-auth/go.mod services/user-auth/go.sum ./services/user-auth/
COPY services/wallet/go.mod services/wallet/go.sum ./services/wallet/

# Copy shared module source
COPY shared/ ./shared/

# Copy service source code
COPY services/user-auth/ ./services/user-auth/

# Set working directory to service
WORKDIR /app/services/user-auth

# Download dependencies with workspace support
ENV GOWORK=/app/go.work
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/services/user-auth/main .

# Change ownership to non-root user
RUN chown appuser:appgroup main

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8084

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8084/health || exit 1

# Run the application
CMD ["./main"]
