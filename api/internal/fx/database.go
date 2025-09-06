package fx

import (
	"api/config"
	"api/internal/database"
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// DatabaseParams holds the parameters needed for database
type DatabaseParams struct {
	fx.In
	Config *config.Config
}

// ProvideDatabase provides the database connection
func ProvideDatabase(lc fx.Lifecycle, params DatabaseParams) (*gorm.DB, error) {
	db, err := database.NewConnection(params.Config)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return database.Close(db)
		},
	})

	return db, nil
}

// DatabaseModule provides the database module
var DatabaseModule = fx.Module("database",
	fx.Provide(ProvideDatabase),
)
