package api

import (
	"go-final-project/pkg/db"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func deleteTaskHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {

	if req.Method != http.MethodDelete {
		logger.Printf("WARN: incorrect request type,")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimSpace(req.URL.Query().Get("id"))

	if id == "" {
		logger.Printf("WARN: the id cannot be empty")
		writeJsonError(res, http.StatusInternalServerError, "Error: the id cannot be empty")
		return
	}

	_, err := strconv.Atoi(id)
	if err != nil {
		logger.Printf("WARN: error for delete action, uncorrect id format, %v", err)
		writeJsonError(res, http.StatusInternalServerError, "Error for delete action, uncorrect id format: "+err.Error())
		return
	}

	err = db.DeleteTask(id)
	if err != nil {
		logger.Printf("WARN: error when delete data, %v", err)
		writeJsonError(res, http.StatusInternalServerError, "Error when delete data: "+err.Error())
		return
	}

	logger.Printf("INFO: task %s completed and deleted", id)
	writeJson(res, http.StatusCreated, map[string]any{})
}
