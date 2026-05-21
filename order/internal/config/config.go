package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP                  httpConfig                  `yaml:"http"`
	Logger                loggerConfig                `yaml:"logger"`
	PG                    pgConfig                    `yaml:"pg"`
	Kafka                 kafkaConfig                 `yaml:"kafka"`
	OrderPaidProducer     orderPaidProducerConfig     `yaml:"order_paid_producer"`
	ShipAssembledConsumer shipAssembledConsumerConfig `yaml:"ship_assembled_consumer_config"`
}

var appConfig *Config

const defaultConfigPath = "order/config.local.yaml"

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
