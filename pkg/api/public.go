package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"tire-pepair-record-service/pkg/db"
)

// normalizeRecord преобразует запись в единый формат для фронтенда
func normalizeRecord(record db.Record) map[string]interface{} {
	normalized := map[string]interface{}{
		"id":      record.ID,
		"date":    record.Date,
		"title":   record.Title,
		"record":  record.Record,
		"comment": record.Comment,
		"status":  record.Status,
	}

	// Генерируем номер талона
	ticketNumber := generateTicketNumber(record.ID, record.Record)
	normalized["ticketNumber"] = ticketNumber

	return normalized
}

// generateTicketNumber генерирует номер талона
func generateTicketNumber(id int64, recordTime *time.Time) string {
	prefix := "О" // Очередь
	if recordTime != nil {
		prefix = "З" // Запись
	}

	// Берем последние 3 цифры ID
	idStr := fmt.Sprintf("%d", id)
	if len(idStr) > 3 {
		idStr = idStr[len(idStr)-3:]
	} else {
		idStr = fmt.Sprintf("%03s", idStr)
	}

	return prefix + idStr
}

// normalizeRecords преобразует массив записей
func normalizeRecords(records []db.Record) []map[string]interface{} {
	normalized := make([]map[string]interface{}, len(records))
	for i, record := range records {
		normalized[i] = normalizeRecord(record)
	}
	return normalized
}

func getTodayRecordsHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodGet {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	records, err := db.GetTodayRecords("")
	if err != nil {
		logger.Printf("ERROR: getting today's records error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: today's records retrieved successfully")
	writeJson(res, http.StatusOK, map[string]any{
		"records": normalizeRecords(records),
	})
}

func getAvailableSlotsHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPost {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var dateReq DateRequest
	if err := json.NewDecoder(req.Body).Decode(&dateReq); err != nil {
		logger.Printf("WARN: unmarshal error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}

	slots, err := db.GetAvailableSlots(dateReq.Date)
	if err != nil {
		logger.Printf("ERROR: getting available slots error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: available slots for %s retrieved successfully", dateReq.Date.Format("2006-01-02"))
	writeJson(res, http.StatusOK, map[string]any{"slots": slots})
}

func getRecordsByDateHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPost {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var dateReq DateRequest
	if err := json.NewDecoder(req.Body).Decode(&dateReq); err != nil {
		logger.Printf("WARN: unmarshal error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}

	records, err := db.GetRecordsByDate(dateReq.Date)
	if err != nil {
		logger.Printf("ERROR: getting records by date error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: records for %s retrieved successfully", dateReq.Date.Format("2006-01-02"))
	writeJson(res, http.StatusOK, map[string]any{"records": records})
}

func addRecordHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPost {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var addReq AddRecordRequest
	if err := json.NewDecoder(req.Body).Decode(&addReq); err != nil {
		logger.Printf("WARN: unmarshal error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}

	// Валидация обязательных полей
	if addReq.Title == "" {
		logger.Printf("WARN: missing required field 'title'")
		writeJsonError(res, http.StatusBadRequest, "Car number is required")
		return
	}

	// Если время указано (предварительная запись), валидируем его
	if addReq.Record != nil {
		if err := db.ValidateRecordTime(*addReq.Record); err != nil {
			logger.Printf("WARN: validation error, %v", err)
			writeJsonError(res, http.StatusBadRequest, err.Error())
			return
		}
	}
	// Если время не указано - это запись в текущую очередь, валидация не нужна

	record := db.Record{
		Title:   addReq.Title,
		Record:  addReq.Record, // может быть nil для текущей очереди
		Comment: addReq.Comment,
		Status:  "wait",
	}

	err := db.AddRecord(record)
	if err != nil {
		logger.Printf("ERROR: adding record error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: record added successfully for car %s", record.Title)
	writeJson(res, http.StatusOK, map[string]any{
		"message": "Record added successfully",
		"success": true,
	})
}
