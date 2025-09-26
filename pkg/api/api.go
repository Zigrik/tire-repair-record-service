package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type AddRecordRequest struct {
	Title   string     `json:"title"`
	Record  *time.Time `json:"record,omitempty"`
	Comment string     `json:"comment"`
}

type UpdateRecordRequest struct {
	ID      int64      `json:"id"`
	Title   string     `json:"title"`
	Record  *time.Time `json:"record,omitempty"`
	Comment string     `json:"comment"`
	Status  string     `json:"status"`
}

type UpdateStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type DateRequest struct {
	Date time.Time `json:"date"`
}

type StatusRequest struct {
	Status string `json:"status"`
}

type PaginationRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Вспомогательная функции для JSON ответов
func writeJson(res http.ResponseWriter, status int, data interface{}) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	if err := json.NewEncoder(res).Encode(data); err != nil {
		http.Error(res, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Вспомогательная функции для JSON ошибок
func writeJsonError(res http.ResponseWriter, status int, message string) {
	writeJson(res, status, map[string]string{"error": message})
}

// Функция инициализации API
func Init(mux *http.ServeMux, logger *log.Logger) {
	mux.HandleFunc("/api/signin", func(res http.ResponseWriter, req *http.Request) {
		signin(res, req, logger)
	})

	// Публичные эндпоинты
	mux.HandleFunc("/api/GetAvailableSlots", func(res http.ResponseWriter, req *http.Request) {
		getAvailableSlotsHandler(res, req, logger)
	})
	mux.HandleFunc("/api/GetRecordsByDate", func(res http.ResponseWriter, req *http.Request) {
		getRecordsByDateHandler(res, req, logger)
	})
	mux.HandleFunc("/api/AddRecord", func(res http.ResponseWriter, req *http.Request) {
		addRecordHandler(res, req, logger)
	})
	mux.HandleFunc("/api/GetTodayRecords", func(res http.ResponseWriter, req *http.Request) {
		getTodayRecordsHandler(res, req, logger)
	})

	getPendingRecords := func(res http.ResponseWriter, req *http.Request) { getPendingRecordsHandler(res, req, logger) }
	getActiveRecords := func(res http.ResponseWriter, req *http.Request) { getActiveRecordsHandler(res, req, logger) }
	updateRecord := func(res http.ResponseWriter, req *http.Request) { updateRecordHandler(res, req, logger) }
	deleteRecord := func(res http.ResponseWriter, req *http.Request) { deleteRecordHandler(res, req, logger) }
	updateRecordStatus := func(res http.ResponseWriter, req *http.Request) { updateRecordStatusHandler(res, req, logger) }
	getAllRecords := func(res http.ResponseWriter, req *http.Request) { getAllRecordsHandler(res, req, logger) }
	getRecordsByStatus := func(res http.ResponseWriter, req *http.Request) { getRecordsByStatusHandler(res, req, logger) }
	getRecordByID := func(res http.ResponseWriter, req *http.Request) { getRecordByIDHandler(res, req, logger) }

	// Защищенные эндпоинты (требуют авторизации)
	mux.HandleFunc("/api/GetPendingRecords", auth(getPendingRecords, logger))
	mux.HandleFunc("/api/GetActiveRecords", auth(getActiveRecords, logger))
	mux.HandleFunc("/api/UpdateRecord", auth(updateRecord, logger))
	mux.HandleFunc("/api/DeleteRecord", auth(deleteRecord, logger))
	mux.HandleFunc("/api/UpdateRecordStatus", auth(updateRecordStatus, logger))
	mux.HandleFunc("/api/GetAllRecords", auth(getAllRecords, logger))
	mux.HandleFunc("/api/GetRecordsByStatus", auth(getRecordsByStatus, logger))
	mux.HandleFunc("/api/GetRecordByID", auth(getRecordByID, logger))
}
