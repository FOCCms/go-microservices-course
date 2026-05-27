package config

type sessionConfig struct {
	TTL string `yaml:"ttl" env:"SESSION_TTL" env-default:"24h"`
}
