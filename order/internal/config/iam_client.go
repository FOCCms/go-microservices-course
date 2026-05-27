package config

import "time"

type iamClientConfig struct {
	Address             string        `yaml:"address" env:"IAM_ADDRESS" env-default:"localhost:50053"`
	KeepaliveTime       time.Duration `yaml:"keepalive_time" env:"IAM_KEEPALIVE_TIME" env-default:"60s"`
	KeepaliveTimeout    time.Duration `yaml:"keepalive_timeout" env:"IAM_KEEPALIVE_TIMEOUT" env-default:"20s"`
	PermitWithoutStream bool          `yaml:"permit_without_stream" env:"IAM_PERMIT_WITHOUT_STREAM" env-default:"true"`
}
