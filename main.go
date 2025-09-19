package main

import (
	"fmt"
	"log"
	"os"
	"tire-pepair-record-service/pkg/api"
	"tire-pepair-record-service/pkg/db"
	"tire-pepair-record-service/server"
)

const (
	portDefault      int    = 7540
	dbDefault        string = "tire_service.db"
	telegramTokenEnv        = "8242376149:AAExwkTZRxjVIC9ztWYoYU2qHwmtlgh_1g0"
)

func main() {
	api.SetPassword()

	logger := log.New(os.Stdout, "server: ", log.LstdFlags)

	err := db.Init(dbDefault, logger)
	if err != nil {
		logger.Fatal("FATAL: error while db load: ", err)
	}
	defer db.CloseDatabase()

	go func() {
		srv := server.StartServer(portDefault, logger)
		if err := srv.HTTPServer.ListenAndServe(); err != nil {
			logger.Fatal("FATAL: error while server start: ", err)
		}
	}()

	// Запуск Telegram бота
	token := telegramTokenEnv
	if token == "" {
		logger.Println("WARN: Telegram bot token not set, bot will not start")
	} else {
		bot, err := NewTelegramBot(token, logger, fmt.Sprintf("http://localhost:%d", portDefault))
		if err != nil {
			logger.Fatalf("FATAL: failed to create telegram bot: %v", err)
		}
		bot.Start()
	}
}
