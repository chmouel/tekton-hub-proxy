package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server          ServerConfig             `mapstructure:"server"`
	ArtifactHub     ArtifactHubConfig        `mapstructure:"artifacthub"`
	CatalogMappings []CatalogMapping         `mapstructure:"catalog_mappings"`
	Logging         LoggingConfig            `mapstructure:"logging"`
	LandingPage     LandingPageConfig        `mapstructure:"landing_page"`
}

type CatalogMapping struct {
	TektonHub    string `mapstructure:"tekton_hub"`
	ArtifactHub  string `mapstructure:"artifact_hub"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type ArtifactHubConfig struct {
	BaseURL    string        `mapstructure:"base_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type LandingPageConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

func Load() (*Config, error) {
	return LoadWithPath("")
}

func LoadWithPath(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}

	// Set environment variable prefix
	viper.SetEnvPrefix("THP")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("artifacthub.base_url", "https://artifacthub.io")
	viper.SetDefault("artifacthub.timeout", "30s")
	viper.SetDefault("artifacthub.max_retries", 3)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("landing_page.enabled", true)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}