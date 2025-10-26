# Migration-only Dockerfile
FROM golang:1.25.1-alpine

# Install git, ca-certificates, and golang-migrate
RUN apk add --no-cache git ca-certificates tzdata
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Set working directory
WORKDIR /migrations

# Copy migrations
COPY api/migrations /migrations

# Default command to run migrations
CMD ["migrate", "-path", "/migrations", "-database", "postgres://postgres:postgres@postgres:5432/aurora_dev?sslmode=disable", "up"]
