package main

import (
	"github.com/zjoart/distributed-notification-system/push-service/internal/config"
	"github.com/zjoart/distributed-notification-system/push-service/internal/database"
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found", logger.WithError(err))
	}

	cfg := config.Load()

	db, err := database.LoadPostgres(cfg.GetPostgresDSN())

	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Fields{
			"error": err.Error(),
		})
	}

	defer db.Close()
	logger.Info("Database connected successfully")

}
