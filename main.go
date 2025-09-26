package main

import (
	"log"
	"os"
	"tire-pepair-record-service/pkg/api"
	"tire-pepair-record-service/pkg/db"
	"tire-pepair-record-service/server"
)

const (
	portDefault int    = 7540
	dbDefault   string = "tire_service.db"
)

func main() {
	api.SetPassword()

	logger := log.New(os.Stdout, "server: ", log.LstdFlags)

	err := db.Init(dbDefault, logger)
	if err != nil {
		logger.Fatal("FATAL: error while db load: ", err)
	}
	defer db.CloseDatabase()

	srv := server.StartServer(portDefault, logger)
	if err := srv.HTTPServer.ListenAndServe(); err != nil {
		logger.Fatal("FATAL: error while server start: ", err)
	}
}
