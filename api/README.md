# Aurora API - Attribute Entity

This implementation provides a complete CRUD system for the Attribute entity, converted from TypeORM to Go with GORM and golang-migrate.

## Features

- **GORM Integration**: Full ORM support with PostgreSQL
- **Database Migrations**: Using golang-migrate CLI for schema management
- **Single Repository Pattern**: Consolidated data access layer
- **Single Service Layer**: Consolidated business logic with validation
- **Type Safety**: Strong typing with Go structs and enums

## Project Structure

```
api/
├── cmd/
│   └── main.go              # Main application entry point
├── config/
│   └── config.go            # Configuration management
├── internal/
│   ├── database/
│   │   └── database.go      # Database connection utilities
│   ├── model/
│   │   └── attribute.go     # Attribute model definition
│   ├── repository/
│   │   └── repository.go    # Consolidated data access layer
│   └── service/
│       └── service.go       # Consolidated business logic layer
└── migrations/
    ├── 001_create_attributes_table.up.sql
    └── 001_create_attributes_table.down.sql
```

## Database Schema

The Attribute entity includes:

- `id`: Primary key (auto-increment)
- `name`: Unique attribute name
- `description`: Text description
- `data_type`: Enum (boolean, string, number, enum)
- `hash_attribute`: Boolean flag for hashing
- `enum_options`: Array of strings for enum values
- `usage_count`: Counter for tracking usage
- `created_at`: Timestamp (auto-managed)
- `updated_at`: Timestamp (auto-managed)

## Setup

1. **Install Dependencies**:
   ```bash
   cd api
   go mod tidy
   ```

2. **Install golang-migrate CLI**:
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

3. **Configure Database**:
   Update `config.yaml` with your PostgreSQL connection details:
   ```yaml
   database:
     host: localhost
     port: 5432
     user: postgres
     password: postgres
     dbname: aurora_dev
     sslmode: disable
   ```

4. **Run Migrations**:
   ```bash
   # From project root
   make migrate-up
   ```

## Usage

### Repository Interface

The `Repository` interface provides all data operations:

- `CreateAttribute(ctx, attribute)` - Create new attribute
- `GetAttributeByID(ctx, id)` - Get attribute by ID
- `GetAttributeByName(ctx, name)` - Get attribute by name
- `GetAllAttributes(ctx, limit, offset)` - Get all attributes with pagination
- `UpdateAttribute(ctx, attribute)` - Update existing attribute
- `DeleteAttribute(ctx, id)` - Delete attribute
- `GetAttributesByDataType(ctx, dataType, limit, offset)` - Filter by data type
- `IncrementAttributeUsageCount(ctx, id)` - Increment usage counter
- `CountAttributes(ctx)` - Get total count

### Service Interface

The `Service` interface provides business logic with validation:

- `CreateAttribute(ctx, request)` - Create with validation
- `UpdateAttribute(ctx, id, request)` - Update with validation
- `GetAllAttributes(ctx, request)` - Get with pagination response
- And more...

### Example Usage

```go
// Create repository and service
repo := repository.New(db)
svc := service.New(repo)

// Create a new attribute
attr, err := svc.CreateAttribute(ctx, &service.CreateAttributeRequest{
    Name:        "user_tier",
    Description: "User subscription tier",
    DataType:    model.DataTypeEnum,
    EnumOptions: []string{"free", "premium", "enterprise"},
})

// Get all attributes
response, err := svc.GetAllAttributes(ctx, &service.GetAllAttributesRequest{
    Limit:  10,
    Offset: 0,
})
```

## Migration Commands

```bash
# Run all pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Rollback all migrations
make migrate-down-all

# Check migration version
make migrate-version

# Force migration version (if stuck)
make migrate-force VERSION=1

# Create new migration
make migrate-create NAME=add_new_table

# Setup development environment
make dev-setup
```

## Data Types

The system supports four data types:

- `boolean`: True/false values
- `string`: Text values
- `number`: Numeric values  
- `enum`: Predefined list of string options

## Validation

- Enum attributes must have `enumOptions` defined
- Attribute names must be unique
- All required fields are validated at service layer

## Architecture

- **Single Repository**: All data access methods in one interface
- **Single Service**: All business logic methods in one interface
- **No Auto-Migration**: Uses golang-migrate CLI for proper schema management
- **Clean Separation**: Repository handles data, Service handles business logic 