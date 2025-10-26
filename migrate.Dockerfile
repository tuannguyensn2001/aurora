# Migration-only Dockerfile
FROM golang:1.25.1-alpine

# Install git, ca-certificates, and golang-migrate
RUN apk add --no-cache git ca-certificates tzdata
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN go install github.com/riverqueue/river/cmd/river@latest

# Set working directory
WORKDIR /migrations

# Copy migrations
COPY api/migrations /migrations

# Create migration script
RUN echo '#!/bin/sh' > /migrations/run-migrations.sh && \
    echo 'echo "Running golang-migrate migrations..."' >> /migrations/run-migrations.sh && \
    echo 'migrate -path /migrations -database "postgres://postgres:postgres@postgres:5432/aurora_dev?sslmode=disable" up' >> /migrations/run-migrations.sh && \
    echo 'echo "Running River migrations..."' >> /migrations/run-migrations.sh && \
    echo 'river migrate-up --database-url "postgres://postgres:postgres@postgres:5432/aurora_dev?sslmode=disable"' >> /migrations/run-migrations.sh && \
    echo 'echo "All migrations completed successfully!"' >> /migrations/run-migrations.sh && \
    chmod +x /migrations/run-migrations.sh

# Default command to run migrations
CMD ["/migrations/run-migrations.sh"]
