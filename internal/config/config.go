package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server ServerConfig `koanf:"server"`
	Cache  CacheConfig  `koanf:"cache"`
	Loader LoaderConfig `koanf:"loader"`
}

type ServerConfig struct {
	Host string `validate:"omitempty" koanf:"host" default:"localhost"`
	Port int    `validate:"omitempty" koanf:"port" default:"8080"`
}

type CacheConfig struct {
	Address  string `validate:"required" koanf:"address"`
	Password string `validate:"omitempty" koanf:"password"`
	DB       int    `validate:"gte=0" koanf:"db"`
}

type LoaderConfig struct {
	MinioConfig     *MinioConfig `validate:"omitempty,dive" koanf:"minio_config"`
	ModuleLoader    string       `validate:"required" koanf:"module_loader"`
	ModuleDirectory string       `validate:"omitempty" koanf:"module_directory"`
}

type MinioConfig struct {
	Endpoint   string `validate:"required" koanf:"endpoint"`
	AccessKey  string `validate:"required" koanf:"access_key"`
	SecretKey  string `validate:"required" koanf:"secret_key"`
	BucketName string `validate:"required" koanf:"bucket_name"`
}

func NewConfig(configFile string) (*Config, error) {
	k := koanf.New(".")

	if _, err := os.Stat(configFile); err == nil {
		if err := k.Load(file.Provider(configFile), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("failed to load configuration from file: %v", err)
		}
	}

	if err := k.Load(env.Provider("", ".", os.Getenv), nil); err != nil {
		return nil, fmt.Errorf("failed to load configuration from environment variables: %v", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %v", err)
	}

	// Fallback values
	if cfg.Server.Host == "" {
		cfg.Server.Host = "localhost"
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &cfg, nil
}

func validateConfig(cfg Config) error {
	validate := validator.New()

	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
