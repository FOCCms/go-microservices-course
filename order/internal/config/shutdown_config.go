package config

import (
	"time"
)

type ShutdownConfig struct {
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"APP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}
