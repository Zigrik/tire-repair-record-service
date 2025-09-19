package api

import (
	"go-final-project/pkg/db"
	"log"
	"net/http"
	"strings"
)

func getTaskHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodGet {
		logger.Printf("WARN: incorrect request type,")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimSpace(req.URL.Query().Get("id"))

	task, err := db.GetTask(id)
	if err != nil {
		logger.Printf("WARN: error when receiving data, %v", err)
		writeJsonError(res, http.StatusInternalServerError, "Error when receiving data: "+err.Error())
		return
	}
	writeJson(res, http.StatusOK, task)
}
