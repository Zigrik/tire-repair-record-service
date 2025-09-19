package api

import (
	"bytes"
	"encoding/json"
	"go-final-project/pkg/db"
	"log"
	"net/http"
	"time"
)

func parseDate(task *db.Task) (time.Time, error) {
	t, err := time.Parse(dateForm, task.Date)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
func addTaskHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
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
			next, err := nextDate(now, task.Date, task.Repeat)
			if err != nil {
				logger.Printf("WARN: date after now error, %v", err)
				writeJsonError(res, http.StatusBadRequest, "Date after now error: "+err.Error())
				return
			}
			task.Date = next
		}
	}
	id, err := db.AddTask(&task)
	if err != nil {
		logger.Printf("WARN: error when adding to the database, %v", err)
		writeJsonError(res, http.StatusBadRequest, "Error when adding to the database: "+err.Error())
		return
	} else {
		logger.Printf("INFO: data with id %d has been added to the database", id)
	}
	writeJson(res, http.StatusCreated, map[string]any{"id": id})
}
