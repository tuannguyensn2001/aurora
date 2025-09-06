server:
	@go run api/cmd/main.go

# Database migration commands using golang-migrate CLI
# Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate-up:
	@migrate -path api/migrations -database "postgres://postgres:postgres@localhost:5432/aurora_dev?sslmode=disable" up

migrate-down:
	@migrate -path api/migrations -database "postgres://postgres:postgres@localhost:5432/aurora_dev?sslmode=disable" down 1

migrate-down-all:
	@migrate -path api/migrations -database "postgres://postgres:postgres@localhost:5432/aurora_dev?sslmode=disable" down

migrate-force:
	@migrate -path api/migrations -database "postgres://postgres:postgres@localhost:5432/aurora_dev?sslmode=disable" force $(VERSION)

migrate-version:
	@migrate -path api/migrations -database "postgres://postgres:postgres@localhost:5432/aurora_dev?sslmode=disable" version

# Create new migration file
migrate-create:
	@migrate create -ext sql -dir api/migrations -seq $(NAME)

# Docker commands
docker-up:
	@echo "Starting PostgreSQL and related services..."
	@docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10
	@echo "PostgreSQL is ready!"

docker-down:
	@echo "Stopping PostgreSQL..."
	@docker-compose down
	@echo "PostgreSQL stopped!"

docker-logs:
	@docker-compose logs -f postgres

docker-clean:
	@echo "Cleaning up Docker resources..."
	@docker-compose down -v
	@docker system prune -f
	@echo "Cleanup complete!"

# Database commands
db-connect:
	@docker exec -it aurora-postgres psql -U postgres -d aurora_dev

db-reset:
	@echo "Resetting database..."
	@make migrate-down-all
	@make migrate-up
	@echo "Database reset complete!"

# Development commands
dev-setup:
	@echo "Setting up development environment..."
	@make docker-up
	@make migrate-up
	@echo "Development setup complete!"
	@echo ""
	@echo "PostgreSQL available at: localhost:5432"

dev-start:
	@make docker-up
	@make server

dev-stop:
	@make docker-down

.PHONY: server migrate-up migrate-down migrate-down-all migrate-force migrate-version migrate-create dev-setup dev-start dev-stop docker-up docker-down docker-logs docker-clean db-connect db-reset