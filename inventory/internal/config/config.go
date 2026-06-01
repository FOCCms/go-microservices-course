package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Stage          string          `yaml:"stage"`
	ServiceVersion string          `yaml:"service_version"`
	GRPC           grpcConfig      `yaml:"grpc"`
	Logger         loggerConfig    `yaml:"logger"`
	PG             pgConfig        `yaml:"pg"`
	IAMClient      iamClientConfig `yaml:"iam_client"`
	ShutdownConfig ShutdownConfig  `yaml:"shutdown_config"`
	OtelConfig     otelConfig      `yaml:"otel"`
	TracingConfig  tracingConfig   `yaml:"tracing_config"`
}

var appConfig *Config

const defaultConfigPath = "inventory/config.local.yaml"

func ResolveConfigPath() string {
	var cfgFlag string
	flag.StringVar(&cfgFlag, "config", "", "путь к YAML-конфигу (например, config.staging.yaml)")
	flag.Parse()

	if cfgFlag != "" {
		return cfgFlag
	}

	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		return envPath
	}
	slog.Warn("config файл не определён, используется " + defaultConfigPath)
	return defaultConfigPath
}

func MustLoad(path string) {
	var cfg Config

	if path != "" {
		if err := cleanenv.ReadConfig(path, &cfg); err != nil {
			panic(fmt.Sprintf("не удалось загрузить конфиг из %q: %q", path, err))
		}
		appConfig = &cfg
		return
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(fmt.Sprintf("не удалось загрузить конфиг из env: %q", err))
	}

	appConfig = &cfg
}

func AppConfig() *Config {
	return appConfig
}
