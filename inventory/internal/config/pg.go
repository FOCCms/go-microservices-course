package config

import "fmt"

type pgConfig struct {
	Host     string `yaml:"host"     env:"POSTGRES_HOST"     env-default:"localhost"`
	Port     string `yaml:"port"     env:"POSTGRES_PORT"     env-default:"5433"`
	Database string `yaml:"database" env:"POSTGRES_DB"       env-default:"inventory-service"`
	User     string `yaml:"user"     env:"POSTGRES_USER"     env-default:"inventory-service-user"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"inventory-service-password"`
	SSLMode  string `yaml:"sslmode"  env:"POSTGRES_SSLMODE"  env-default:"disable"`
}

func (c *pgConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.User, c.Password, c.SSLMode,
	)
}
