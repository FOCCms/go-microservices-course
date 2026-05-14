package main

import (
	"context"
	"log/slog"

	"github.com/FOCCms/go-microservices-course/assembly/internal/app"
	"github.com/FOCCms/go-microservices-course/assembly/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	// .env опционален — ошибка загрузки допустима.
	err := godotenv.Load("assembly.env")
	if err != nil {
		slog.Warn("не удалось загрузить .env конфигурацию", "error", err)
	}

	config.MustLoad(config.ResolveConfigPath())

	a := app.New(context.Background())

	if err := a.Run(); err != nil {
		slog.Error("ошибка при работе приложения", "error", err)
	}

}
