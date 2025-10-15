# Aurora

**Open Source Feature Flagging and A/B Testing Platform**

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-316192?style=flat&logo=postgresql)](https://www.postgresql.org)

---

## 🚀 Quick Start

Get up and running in under 2 minutes:

```bash
git clone https://github.com/your-org/aurora.git
cd aurora
make dev-setup
make server
```

Then visit `http://localhost:9999` to access the API.

## 📖 Table of Contents

- [Philosophy](#philosophy)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [SDK Usage](#sdk-usage)
- [Configuration](#configuration)
- [Development](#development)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Philosophy

The top 1% of companies invest thousands of engineering hours building sophisticated feature flagging and experimentation platforms in-house. The remaining 99% either pay premium prices for SaaS solutions or struggle with fragmented open-source libraries.

**Aurora bridges this gap.** We provide the power and flexibility of an enterprise-grade experimentation platform without the overhead of building it yourself or the recurring costs of third-party solutions.

### Why Aurora?

- **🔓 Open Source**: Complete control over your experimentation infrastructure
- **⚡ High Performance**: Built in Go for speed and efficiency
- **🏗️ Modern Architecture**: Clean, maintainable codebase using industry best practices
- **📊 Enterprise Ready**: Designed for scale with support for millions of evaluations
- **🔌 Easy Integration**: Simple SDKs and REST APIs for seamless adoption

## ✨ Features

### Core Capabilities

- **🚩 Feature Flags**: Advanced targeting, gradual rollouts, and kill switches
- **🧪 A/B Testing**: Run experiments with sophisticated targeting rules
- **🎯 Audience Segmentation**: Define user segments based on custom attributes
- **📈 Dynamic Parameters**: Change feature configurations without deployments
- **🔄 Real-time Updates**: SDK automatically fetches latest configurations
- **💾 Local Caching**: Ultra-fast evaluations with BadgerDB-backed storage
- **☁️ S3 Distribution**: Optional CDN-like configuration delivery via AWS S3
- **🔗 API-First Design**: RESTful APIs for all platform operations

### Technical Features

- **🎲 Consistent Hashing**: Deterministic user bucketing with MurmurHash3
- **🔐 Type-Safe Evaluation**: Strong typing for parameters (string, number, boolean, JSON)
- **⚙️ Flexible Configuration**: YAML-based configuration with environment overrides
- **📦 Zero Dependencies**: Standalone Go binaries, no external runtime required
- **🔄 Background Sync**: Automatic configuration updates with configurable refresh rates
- **🛡️ Error Handling**: Comprehensive error types with graceful fallbacks

## 🏛️ Architecture

Aurora follows a modern microservices architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Application                       │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │              Aurora SDK (Go)                       │    │
│  │  • Local BadgerDB Cache                            │    │
│  │  • Evaluation Engine                               │    │
│  │  • Background Sync                                 │    │
│  └────────────────────────────────────────────────────┘    │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        │ REST API / S3
                        │
┌───────────────────────▼──────────────────────────────────────┐
│                     Aurora API Server                        │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Gin HTTP   │  │  GORM ORM    │  │  River Jobs  │     │
│  │   Router     │  │              │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Experiments │  │  Parameters  │  │  Segments    │     │
│  │  Service     │  │  Service     │  │  Service     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└───────────────────────┬──────────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
  ┌──────────┐   ┌──────────┐   ┌──────────┐
  │PostgreSQL│   │  AWS S3  │   │  River   │
  │          │   │  Bucket  │   │  Queue   │
  └──────────┘   └──────────┘   └──────────┘
```

### Components

- **API Server**: Go-based REST API built with Gin framework
- **SDK**: Lightweight Go client library with local caching
- **Database**: PostgreSQL for persistence (experiments, parameters, segments)
- **Cache**: BadgerDB for SDK-side caching
- **Job Queue**: River for background task processing
- **Storage**: Optional S3 for configuration distribution

## 🎬 Getting Started

### Prerequisites

- Go 1.24+ 
- Docker & Docker Compose
- PostgreSQL 15+ (via Docker)
- Make (optional, for convenience commands)

### Installation

#### Option 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-org/aurora.git
cd aurora

# Start PostgreSQL
make docker-up

# Run migrations
make migrate-up

# Start the API server
make server
```

#### Option 2: Manual Setup

```bash
# Install dependencies
cd api
go mod download

cd ../sdk
go mod download

# Configure database connection
cp config.tmp.yaml config.yaml
# Edit config.yaml with your settings

# Run migrations
migrate -path api/migrations -database "postgres://postgres:postgres@localhost:5432/aurora_dev?sslmode=disable" up

# Start the server
cd api
go run cmd/main.go -config ../config.yaml
```

### Database Setup

The project uses `golang-migrate` for database migrations. Install it:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Available migration commands:

```bash
make migrate-up         # Apply all migrations
make migrate-down       # Rollback one migration
make migrate-down-all   # Rollback all migrations
make migrate-create NAME=your_migration_name  # Create new migration
```

## 💻 SDK Usage

### Installation

```bash
go get github.com/your-org/aurora/sdk
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/your-org/aurora/sdk"
)

func main() {
    // Initialize the client
    client, err := sdk.NewClient(
        sdk.ClientOptions{
            EndpointURL:  "http://localhost:9999",
            S3BucketName: "your-s3-bucket", // optional
        },
        sdk.WithRefreshRate(30*time.Second),
        sdk.WithInMemoryOnly(true), // Use in-memory cache only
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start the client (begins background sync)
    err = client.Start(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()
    
    // Create user attributes for targeting
    userAttrs := sdk.NewAttribute().
        SetString("user_id", "user_12345").
        SetString("country", "US").
        SetNumber("age", 28).
        SetBoolean("premium", true)
    
    // Evaluate a parameter
    result := client.EvaluateParameter(
        context.Background(),
        "welcome_message",
        userAttrs,
    )
    
    if result.HasError() {
        log.Printf("Error evaluating parameter: %v", result.Error())
        // Use default value
        message := result.AsString("Hello, World!")
        fmt.Println(message)
    } else {
        message := result.AsString("Hello, World!")
        fmt.Println(message)
    }
}
```

### Advanced Configuration

```go
// With custom refresh rate and logging
client, err := sdk.NewClient(
    sdk.ClientOptions{
        EndpointURL:  "https://your-aurora-instance.com",
        S3BucketName: "prod-experiments",
    },
    sdk.WithRefreshRate(60*time.Second),
    sdk.WithLogLevel(slog.LevelInfo),
    sdk.WithPath("./experiment-cache"),
    sdk.WithEnableS3(true),
    sdk.WithOnEvaluate(func(source, paramName string, attrs *sdk.Attribute, value *string, err error) {
        // Custom evaluation callback for analytics
        log.Printf("Evaluated %s from %s: %v", paramName, source, value)
    }),
)
```

### Working with Different Data Types

```go
// String parameters
welcomeMsg := client.EvaluateParameter(ctx, "welcome_message", attrs).
    AsString("Hello!")

// Number parameters
maxRetries := client.EvaluateParameter(ctx, "max_retries", attrs).
    AsInt(3)

timeout := client.EvaluateParameter(ctx, "timeout_seconds", attrs).
    AsFloat64(30.0)

// Boolean parameters
featureEnabled := client.EvaluateParameter(ctx, "new_feature_enabled", attrs).
    AsBool(false)

// JSON parameters
var config map[string]interface{}
err := client.EvaluateParameter(ctx, "feature_config", attrs).
    AsJSON(&config)
```

### Error Handling

```go
result := client.EvaluateParameter(ctx, "my_param", attrs)

if result.HasError() {
    switch result.Error().(type) {
    case *sdk.ParameterNotFoundError:
        log.Println("Parameter not found, using default")
    case *sdk.NetworkError:
        log.Println("Network error, using cached value")
    case *sdk.ConfigurationError:
        log.Println("Configuration error")
    default:
        log.Printf("Unknown error: %v", result.Error())
    }
}

// Always safe to call with default value
value := result.AsString("default_value")
```

## ⚙️ Configuration

Aurora uses YAML configuration files. Create a `config.yaml` in your project root:

```yaml
service:
  name: aurora-api
  env: production
  port: 9999

logging:
  level: info  # debug, info, warn, error

database:
  host: localhost
  port: 5432
  user: postgres
  password: your_secure_password
  dbname: aurora_prod
  sslmode: require
  
s3:
  enable: true
  bucketName: prod-aurora-experiments
  accessKey: YOUR_AWS_ACCESS_KEY
  secretKey: YOUR_AWS_SECRET_KEY
```

### Environment Variables

You can override configuration values with environment variables:

```bash
export SERVICE_PORT=8080
export DATABASE_HOST=prod-db.example.com
export S3_ENABLE=false
```

## 🛠️ Development

### Project Structure

```
aurora/
├── api/                          # API Server
│   ├── cmd/
│   │   └── main.go              # Application entry point
│   ├── config/                   # Configuration management
│   ├── internal/
│   │   ├── app/                 # Application utilities
│   │   ├── constant/            # Constants
│   │   ├── database/            # Database setup
│   │   ├── dto/                 # Data Transfer Objects
│   │   ├── fx/                  # Dependency injection (Uber FX)
│   │   ├── handler/             # HTTP handlers
│   │   ├── mapper/              # Data mappers
│   │   ├── model/               # Database models
│   │   ├── repository/          # Data access layer
│   │   ├── router/              # HTTP routing
│   │   ├── service/             # Business logic
│   │   └── workers/             # Background jobs
│   └── migrations/              # Database migrations
│
├── sdk/                          # Go SDK
│   ├── client.go                # Main SDK client
│   ├── engine.go                # Evaluation engine
│   ├── storage.go               # BadgerDB storage
│   ├── attribute.go             # User attributes
│   ├── value.go                 # Evaluation results
│   └── errors.go                # Error types
│
├── config.yaml                   # Configuration file
├── docker-compose.yaml          # Docker setup
├── Makefile                     # Development commands
└── README.md                    # This file
```

### Available Make Commands

```bash
# Development
make dev-setup      # Setup development environment
make dev-start      # Start development server
make dev-stop       # Stop development server
make server         # Run API server

# Database
make migrate-up     # Apply migrations
make migrate-down   # Rollback last migration
make migrate-create NAME=migration_name  # Create new migration
make db-connect     # Connect to PostgreSQL
make db-reset       # Reset database

# Docker
make docker-up      # Start PostgreSQL
make docker-down    # Stop PostgreSQL
make docker-logs    # View PostgreSQL logs
make docker-clean   # Clean Docker resources
```

### Running Tests

```bash
# Test SDK
cd sdk
go test -v ./...

# Test API
cd api
go test -v ./...
```

### Creating a New Migration

```bash
make migrate-create NAME=add_new_feature

# This creates two files:
# api/migrations/{timestamp}_add_new_feature.up.sql
# api/migrations/{timestamp}_add_new_feature.down.sql
```

## 📚 API Documentation

### Endpoints

#### SDK Endpoints

```
POST /api/v1/sdk/parameters
POST /api/v1/sdk/experiments
GET  /api/v1/sdk/metadata
```

#### Management Endpoints

```
# Experiments
GET    /api/v1/experiments
POST   /api/v1/experiments
GET    /api/v1/experiments/:id
PUT    /api/v1/experiments/:id
DELETE /api/v1/experiments/:id

# Parameters
GET    /api/v1/parameters
POST   /api/v1/parameters
GET    /api/v1/parameters/:id
PUT    /api/v1/parameters/:id
DELETE /api/v1/parameters/:id

# Segments
GET    /api/v1/segments
POST   /api/v1/segments
GET    /api/v1/segments/:id
PUT    /api/v1/segments/:id
DELETE /api/v1/segments/:id

# Attributes
GET    /api/v1/attributes
POST   /api/v1/attributes
GET    /api/v1/attributes/:id
PUT    /api/v1/attributes/:id
DELETE /api/v1/attributes/:id
```

### Example API Requests

#### Create an Experiment

```bash
curl -X POST http://localhost:9999/api/v1/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Homepage Redesign",
    "description": "Testing new homepage layout",
    "status": "running",
    "hashAttributeId": 1,
    "variants": [
      {
        "name": "Control",
        "trafficAllocation": 50,
        "parameters": [
          {
            "parameterName": "hero_title",
            "rolloutValue": "Welcome to our site"
          }
        ]
      },
      {
        "name": "Variant A",
        "trafficAllocation": 50,
        "parameters": [
          {
            "parameterName": "hero_title",
            "rolloutValue": "Discover amazing features"
          }
        ]
      }
    ]
  }'
```

#### Create a Parameter

```bash
curl -X POST http://localhost:9999/api/v1/parameters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "max_upload_size",
    "dataType": "number",
    "defaultValue": "10485760",
    "description": "Maximum file upload size in bytes"
  }'
```

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** and write tests
4. **Run tests**: `go test ./...`
5. **Commit your changes**: `git commit -m 'Add amazing feature'`
6. **Push to the branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new features
- Update documentation for API changes
- Use meaningful commit messages
- Keep pull requests focused and atomic

### Code Style

This project follows standard Go conventions:

- Use `gofmt` for formatting
- Follow effective Go guidelines
- Write idiomatic Go code
- Add comments for exported functions

## 📋 Roadmap

- [ ] Web UI for experiment management
- [ ] Multi-language SDK support (JavaScript, Python, Ruby, Java)
- [ ] Advanced analytics and reporting
- [ ] Webhook support for real-time notifications
- [ ] Feature flag scheduling
- [ ] Gradual rollout controls
- [ ] Multi-variate testing (MVT)
- [ ] Integration with popular analytics platforms
- [ ] Audit logs and change history
- [ ] Role-based access control (RBAC)

## 🔐 Security

If you discover a security vulnerability, please email security@your-domain.com instead of opening a public issue.

## 📄 License

Aurora is open source software licensed under the [MIT License](LICENSE).

## 🙏 Acknowledgments

Aurora draws inspiration from industry-leading platforms:

- [GrowthBook](https://github.com/growthbook/growthbook) - Open source feature flagging and A/B testing
- [Statsig](https://www.statsig.com/) - Modern experimentation platform
- [LaunchDarkly](https://launchdarkly.com/) - Feature management platform

## 💬 Support

- 📖 [Documentation](https://docs.your-aurora-instance.com) (coming soon)
- 💬 [Community Slack](https://join.slack.com/your-workspace) (coming soon)
- 🐛 [Issue Tracker](https://github.com/your-org/aurora/issues)
- 📧 Email: support@your-domain.com

## 📊 Stats

![GitHub stars](https://img.shields.io/github/stars/your-org/aurora?style=social)
![GitHub forks](https://img.shields.io/github/forks/your-org/aurora?style=social)
![GitHub issues](https://img.shields.io/github/issues/your-org/aurora)
![GitHub pull requests](https://img.shields.io/github/issues-pr/your-org/aurora)

---

**Built with ❤️ by the Aurora Team**

[Website](https://your-aurora-instance.com) • [Documentation](https://docs.your-aurora-instance.com) • [Blog](https://blog.your-aurora-instance.com)

