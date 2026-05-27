package config

import "time"

type paymentClientConfig struct {
	Address             string        `yaml:"address" env:"PAYMENT_ADDRESS" env-default:"localhost:50052"`
	KeepaliveTime       time.Duration `yaml:"keepalive_time" env:"PAYMENT_KEEPALIVE_TIME" env-default:"30s"`
	KeepaliveTimeout    time.Duration `yaml:"keepalive_timeout" env:"PAYMENT_KEEPALIVE_TIMEOUT" env-default:"3s"`
	PermitWithoutStream bool          `yaml:"permit_without_stream" env:"PAYMENT_PERMIT_WITHOUT_STREAM" env-default:"true"`
}
