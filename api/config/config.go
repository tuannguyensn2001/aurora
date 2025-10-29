package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
	Service struct {
		Name string `yaml:"name"`
		Env  string `yaml:"env"`
		Port int    `yaml:"port"`
	} `yaml:"service"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
	S3 struct {
		Enable     bool   `yaml:"enable"`
		BucketName string `yaml:"bucketName"`
	} `yaml:"s3"`
	OAuth struct {
		Google struct {
			ClientID     string `yaml:"clientId"`
			ClientSecret string `yaml:"clientSecret"`
			RedirectURL  string `yaml:"redirectUrl"`
		} `yaml:"google"`
		AllowedDomains []string `yaml:"allowedDomains"` // List of allowed email domains (e.g., ["example.com", "company.org"])
	} `yaml:"oauth"`
	JWT struct {
		Secret     string `yaml:"secret"`
		ExpireHour int    `yaml:"expireHour"`
	} `yaml:"jwt"`
	Solver struct {
		EndpointURL string `yaml:"endpointUrl"`
	} `yaml:"solver"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Service.Env == "development"
}
