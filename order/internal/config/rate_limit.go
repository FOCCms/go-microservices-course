package config

type rateLimitConfig struct {
	Address string `yaml:"redis_address"`
	Rate    int    `yaml:"rate"`
	Burst   int    `yaml:"burst"`
}
