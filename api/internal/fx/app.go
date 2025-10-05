package fx

import (
	"net/http"
	"sdk"

	"go.uber.org/fx"
)

// NewApp creates a new FX application with all modules
func NewApp(configPath string) *fx.App {
	return fx.New(
		// Provide config path as a named dependency
		fx.Provide(
			fx.Annotate(
				func() string { return configPath },
				fx.ResultTags(`name:"config_path"`),
			),
		),

		// Include all modules
		S3Module,
		ConfigModule,
		LoggerModule,
		DatabaseModule,
		RepositoryModule,
		ServiceModule,
		HandlerModule,
		ServerModule,
		WorkerModule,
		RiverModule,
		SDKModule,

		// Invoke server to ensure it starts
		fx.Invoke(func(*http.Server, sdk.Client) {}),
	)
}
