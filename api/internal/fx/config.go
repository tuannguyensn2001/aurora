package fx

import (
	"api/config"
	"flag"

	"go.uber.org/fx"
)

// ConfigParams holds the parameters needed for config
type ConfigParams struct {
	fx.In
	ConfigPath string `name:"config_path"`
}

// ProvideConfig provides the config instance
func ProvideConfig(params ConfigParams) (*config.Config, error) {
	return config.Load(params.ConfigPath)
}

// ConfigModule provides the config module
var ConfigModule = fx.Module("config",
	fx.Provide(ProvideConfig),
)

// ProvideConfigPath provides the config path from command line flags
func ProvideConfigPath() string {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()
	return *configPath
}
