# Aurora—Open Source Feature Flagging and A/B Testing Platform

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-316192?style=flat&logo=postgresql)](https://www.postgresql.org)
[![GitHub stars](https://img.shields.io/github/stars/your-org/aurora?style=social)](https://github.com/your-org/aurora)

## Introduction

Aurora is a durable experimentation platform that enables developers to build scalable A/B testing and feature flagging systems without sacrificing performance or reliability. The Aurora server executes feature evaluations and experiment assignments in a resilient manner that automatically handles configuration updates, network failures, and provides consistent user experiences.

Aurora is built with modern Go architecture, providing enterprise-grade experimentation capabilities with the simplicity of open source. It combines the power of sophisticated targeting rules with high-performance local caching to deliver sub-millisecond evaluation times.

### Why Aurora?

- **🔓 Open Source**: Complete control over your experimentation infrastructure without vendor lock-in
- **⚡ High Performance**: Built in Go for speed and efficiency, handling millions of evaluations per second
- **🏗️ Modern Architecture**: Clean, maintainable codebase using industry best practices
- **📊 Enterprise Ready**: Designed for scale with support for complex targeting and segmentation
- **🔌 Easy Integration**: Simple SDKs for seamless adoption across your stack

## 🚀 Quick Start

### Download and Start Aurora Server Locally

Execute the following commands to start Aurora with all dependencies:

```bash
# Clone the repository
git clone https://github.com/your-org/aurora.git
cd aurora

# Start PostgreSQL and run migrations
make dev-setup

# Start the Aurora API server
make server
```

The Aurora API will be available at `http://localhost:9999`.

### Use the SDK

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
    // Initialize Aurora SDK
    client, err := sdk.NewClient(sdk.ClientOptions{
        EndpointURL: "http://localhost:9999",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Start the client
    err = client.Start(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()
    
    // Create user attributes
    userAttrs := sdk.NewAttribute().
        SetString("user_id", "user_12345").
        SetString("country", "US")
    
    // Evaluate parameter
    result := client.EvaluateParameter(
        context.Background(),
        "welcome_message",
        userAttrs,
    )
    
    message := result.AsString("Hello, World!")
    fmt.Println(message) // Output: Hello, World!
}
```

## 📖 Table of Contents

- [Features](#-features)
- [Getting Started](#-getting-started)
- [SDK Usage](#-sdk-usage)
- [License](#-license)

## ✨ Features

- **🚩 Feature Flags**: Advanced targeting, gradual rollouts, and kill switches
- **🧪 A/B Testing**: Run experiments with sophisticated targeting rules
- **🎯 Audience Segmentation**: Define user segments based on custom attributes
- **📈 Dynamic Parameters**: Change feature configurations without deployments
- **🔄 Real-time Updates**: SDK automatically fetches latest configurations
- **💾 Local Caching**: Ultra-fast evaluations with BadgerDB-backed storage
- **☁️ S3 Distribution**: Optional CDN-like configuration delivery via AWS S3

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

The Aurora SDK provides a lightweight Go client library for evaluating feature flags and A/B test parameters. It features local caching, automatic configuration updates, and type-safe evaluation methods.

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

### Real-World Examples

#### E-commerce Feature Flags

```go
// Check if new checkout flow is enabled for this user
isNewCheckoutEnabled := client.EvaluateParameter(ctx, "new_checkout_enabled", userAttrs).
    AsBool(false)

if isNewCheckoutEnabled {
    // Show new checkout UI
    renderNewCheckout()
} else {
    // Show legacy checkout
    renderLegacyCheckout()
}

// Get dynamic pricing multiplier
pricingMultiplier := client.EvaluateParameter(ctx, "pricing_multiplier", userAttrs).
    AsNumber(1.0)

finalPrice := basePrice * pricingMultiplier
```

#### Content Personalization

```go
// Get personalized content based on user segment
contentTheme := client.EvaluateParameter(ctx, "content_theme", userAttrs).
    AsString("default")

switch contentTheme {
case "dark":
    renderDarkTheme()
case "minimal":
    renderMinimalTheme()
default:
    renderDefaultTheme()
}

// Get user-specific feature configurations
var featureConfig map[string]interface{}
err := client.EvaluateParameter(ctx, "feature_config", userAttrs).
    AsJSON(&featureConfig)

if err == nil {
    maxUploadSize := featureConfig["max_upload_size"].(float64)
    enableNotifications := featureConfig["notifications"].(bool)
}
```

#### A/B Testing Experiments

```go
// Run homepage redesign experiment
homepageVersion := client.EvaluateParameter(ctx, "homepage_version", userAttrs).
    AsString("control")

switch homepageVersion {
case "variant_a":
    renderHomepageVariantA()
case "variant_b":
    renderHomepageVariantB()
default:
    renderHomepageControl()
}

// Track experiment exposure for analytics
if homepageVersion != "control" {
    analytics.Track("experiment_exposure", map[string]interface{}{
        "experiment": "homepage_redesign",
        "variant": homepageVersion,
        "user_id": userAttrs.Get("user_id"),
    })
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

## 📄 License

Aurora is open source software licensed under the [MIT License](LICENSE).

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

