package api

import (
	"go-final-project/pkg/db"
	"log"
	"net/http"
	"strings"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

const tasksOnPage int = 50

func tasksHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodGet {
		logger.Printf("WARN: incorrect request type,")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	search := strings.TrimSpace(req.URL.Query().Get("search"))

	tasks, err := db.Tasks(tasksOnPage, search, dateForm)
	if err != nil {
		logger.Printf("WARN: error when receiving data, %v", err)
		writeJsonError(res, http.StatusInternalServerError, "Error when receiving data: "+err.Error())
		return
	}
	writeJson(res, http.StatusOK, TasksResp{Tasks: tasks})
}
