package config

import (
	"net"
	"time"
)

type grpcConfig struct {
	Host                  string        `yaml:"host" env:"GRPC_HOST" env-default:"localhost"`
	Port                  string        `yaml:"port" env:"GRPC_PORT" env-default:"50053"`
	MaxConnectionIdle     time.Duration `yaml:"max_connection_idle" env:"GRPC_MAX_CONNECTION_IDLE" env-default:"15m"`
	MaxConnectionAge      time.Duration `yaml:"max_connection_age" env:"GRPC_MAX_CONNECTION_AGE" env-default:"30m"`
	MaxConnectionAgeGrace time.Duration `yaml:"max_connection_age_grace" env:"GRPC_MAX_CONNECTION_AGE_GRACE" env-default:"10s"`
	KeepaliveTime         time.Duration `yaml:"keepalive_time" env:"GRPC_KEEPALIVE_TIME" env-default:"5m"`
	KeepaliveTimeout      time.Duration `yaml:"keepalive_timeout" env:"GRPC_KEEPALIVE_TIMEOUT" env-default:"5s"`
	MinPingInterval       time.Duration `yaml:"min_ping_interval" env:"GRPC_MIN_PING_INTERVAL" env-default:"5m"`
}

func (c *grpcConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}
