package api

import (
	"go-final-project/pkg/db"
	"log"
	"net/http"
	"strings"
	"time"
)

func taskDoneHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {

	if req.Method != http.MethodPost {
		logger.Printf("WARN: incorrect request type,")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := strings.TrimSpace(req.URL.Query().Get("id"))

	task, err := db.GetTask(id)
	if err != nil {
		logger.Printf("WARN: error when receiving data for done action, %v", err)
		writeJsonError(res, http.StatusInternalServerError, "Error when receiving data for done action: "+err.Error())
		return
	}

	if len(task.Repeat) == 0 {
		err = db.DeleteTask(id)
		if err != nil {
			logger.Printf("WARN: error when delete data, %v", err)
			writeJsonError(res, http.StatusInternalServerError, "Error when delete data: "+err.Error())
			return
		}
		logger.Printf("WARN: task %s completed and deleted", id)
		writeJson(res, http.StatusOK, map[string]any{})
		return
	} else {
		dateNew, err := nextDate(time.Now().AddDate(0, 0, 1), task.Date, task.Repeat)
		if err != nil {
			logger.Printf("WARN: error NextDay func, %v", err)
			writeJsonError(res, http.StatusInternalServerError, "Error NextDay func: "+err.Error())
			return
		}
		task.Date = dateNew

		err = db.UpdateTask(task)
		if err != nil {
			logger.Printf("WARN: error when updating row, %v", err)
			writeJsonError(res, http.StatusBadRequest, "Error when updating data in the database: "+err.Error())
			return
		}
	}

	logger.Printf("INFO: data with id %s has been updating in the database", task.ID)
	writeJson(res, http.StatusOK, map[string]any{})
}
