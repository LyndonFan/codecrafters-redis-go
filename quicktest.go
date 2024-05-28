package main

import (
    customLogger "github.com/codecrafters-io/redis-starter-go/app/logger"
)

func main() {
	logger := customLogger.NewLogger(1234, customLogger.LOG_LEVEL_INFO)
	logger.Info("hello")
}
