package main

import (
	"github.com/zjoart/distributed-notification-system/push-service/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found", logger.WithError(err))
	}

	// cfg := config.Load()

}
