package fx

import (
	"api/internal/repository"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// RepositoryParams holds the parameters needed for repository
type RepositoryParams struct {
	fx.In
	DB *gorm.DB
}

// ProvideRepository provides the repository instance
func ProvideRepository(params RepositoryParams) repository.Repository {
	return repository.New(params.DB)
}

// RepositoryModule provides the repository module
var RepositoryModule = fx.Module("repository",
	fx.Provide(ProvideRepository),
)
