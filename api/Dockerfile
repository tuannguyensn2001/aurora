# Multi-stage build for Aurora API
FROM golang:1.25.1-alpine AS builder

# Install git and ca-certificates for dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go work files
COPY go.work go.work.sum ./

# Copy go mod files for both modules (needed for workspace)
COPY api/go.mod api/go.sum ./api/
COPY sdk/go.mod sdk/go.sum ./sdk/

# Download dependencies
RUN go work sync

# Copy source code
COPY api/ ./api/
COPY sdk/ ./sdk/

# Build the API binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./api/cmd/main.go

# Final stage
FROM alpine:3.19

# Install ca-certificates and wget for health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 9999

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9999/health || exit 1

# Run the application
CMD ["./main", "-config", "config.yaml"]
