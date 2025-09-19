package api

import (
	"bytes"
	"encoding/json"
	"go-final-project/pkg/db"
	"log"
	"net/http"
	"time"
)

func putTaskHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPut {
		logger.Printf("WARN: incorrect request type,")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var task db.Task
	var buf bytes.Buffer

	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		logger.Printf("WARN: request reading error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		logger.Printf("WARN: unmarshal error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}
	if task.Title == "" {
		logger.Println("WARN: the title shelf should not be empty")
		writeJsonError(res, http.StatusBadRequest, "The title shelf should not be empty")
		return
	}

	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(dateForm)
	}

	t, err := parseDate(&task)
	if err != nil {
		logger.Printf("WARN: checkdate error, %v", err)
		writeJsonError(res, http.StatusBadRequest, "Checkdate error: "+err.Error())
		return
	}

	if AfterNow(time.Now(), t) {
		if len(task.Repeat) == 0 {
			task.Date = now.Format(dateForm)
		} else {
			next, err := nextDate(now.AddDate(0, 0, 1), task.Date, task.Repeat)
			if err != nil {
				logger.Printf("WARN: date after now error, %v", err)
				writeJsonError(res, http.StatusBadRequest, "Date after now error: "+err.Error())
				return
			}
			task.Date = next
		}
	}

	_, err = parseDate(&task)
	if err != nil {
		logger.Printf("WARN: checkdate error, %v", err)
		writeJsonError(res, http.StatusBadRequest, "Checkdate error: "+err.Error())
		return
	}

	_, err = db.GetTask(task.ID)
	if err != nil {
		logger.Printf("WARN: uncorrect ID for update task, %v", err)
		writeJsonError(res, http.StatusBadRequest, "Check task error: "+err.Error())
		return
	}

	err = db.UpdateTask(&task)
	if err != nil {
		logger.Printf("WARN: error when updating row, %v", err)
		writeJsonError(res, http.StatusBadRequest, "Error when updating data in the database: "+err.Error())
		return
	}

	//request a task from the database after the change to ensure that the data in the database is correct
	taskResult, err := db.GetTask(task.ID)
	if err != nil {
		logger.Printf("WARN: uncorrect ID for update task, %v", err)
		writeJsonError(res, http.StatusBadRequest, "Check task error: "+err.Error())
		return
	}

	logger.Printf("INFO: data with id %s has been updating in the database", task.ID)
	writeJson(res, http.StatusOK, taskResult)
}
