package main

import (
	"context"
	"log/slog"

	"github.com/joho/godotenv"

	"github.com/FOCCms/go-microservices-course/iam/internal/app"
	"github.com/FOCCms/go-microservices-course/iam/internal/config"
)

func main() {
	// .env опционален — ошибка загрузки допустима.
	err := godotenv.Load("payment.env")
	if err != nil {
		slog.Warn("не удалось загрузить .env конфигурацию", "error", err)
	}

	config.MustLoad(config.ResolveConfigPath())

	a, err := app.New(context.Background())
	if err != nil {
		slog.Error("ошибка запуска приложения", "error", err)
		return
	}

	if err := a.Run(); err != nil {
		slog.Error("ошибка при работе приложения", "error", err)
	}
}
