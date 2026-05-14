package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Kafka                 kafkaConfig                 `yaml:"kafka"`
	OrderPaidConsumer     orderPaidConsumerConfig     `yaml:"order_paid_consumer"`
	ShipAssembledProducer shipAssembledProducerConfig `yaml:"ship_assembled_producer"`
}

var appConfig *Config

const defaultConfigPath = "assembly/config.local.yaml"

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
