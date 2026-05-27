package config

import "time"

type inventoryClientConfig struct {
	Address             string        `yaml:"address" env:"INVENTORY_ADDRESS" env-default:"localhost:50051"`
	KeepaliveTime       time.Duration `yaml:"keepalive_time" env:"INVENTORY_KEEPALIVE_TIME" env-default:"30s"`
	KeepaliveTimeout    time.Duration `yaml:"keepalive_timeout" env:"INVENTORY_KEEPALIVE_TIMEOUT" env-default:"3s"`
	PermitWithoutStream bool          `yaml:"permit_without_stream" env:"INVENTORY_PERMIT_WITHOUT_STREAM" env-default:"true"`
}
