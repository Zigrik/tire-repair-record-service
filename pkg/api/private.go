package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"tire-pepair-record-service/pkg/db"
)

func getPendingRecordsHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodGet {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	records, err := db.GetPendingRecords()
	if err != nil {
		logger.Printf("ERROR: getting pending records error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: pending records retrieved successfully")
	writeJson(res, http.StatusOK, map[string]any{"records": records})
}

func getActiveRecordsHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodGet {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	records, err := db.GetActiveRecords()
	if err != nil {
		logger.Printf("ERROR: getting active records error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: active records retrieved successfully")
	writeJson(res, http.StatusOK, map[string]any{"records": records})
}

func updateRecordHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPut {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var updateReq UpdateRecordRequest
	if err := json.NewDecoder(req.Body).Decode(&updateReq); err != nil {
		logger.Printf("WARN: unmarshal error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}

	record := db.Record{
		ID:      updateReq.ID,
		Title:   updateReq.Title,
		Record:  updateReq.Record,
		Comment: updateReq.Comment,
		Status:  updateReq.Status,
	}

	if record.Record != nil {
		if err := db.ValidateRecordTime(*record.Record); err != nil {
			logger.Printf("WARN: validation error, %v", err)
			writeJsonError(res, http.StatusBadRequest, err.Error())
			return
		}
	}

	err := db.UpdateRecord(record.ID, record)
	if err != nil {
		logger.Printf("ERROR: updating record error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: record %d updated successfully", record.ID)
	writeJson(res, http.StatusOK, map[string]any{"message": "Record updated successfully"})
}

func deleteRecordHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodDelete {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	recordIDStr := req.URL.Query().Get("id")
	if recordIDStr == "" {
		logger.Printf("WARN: missing record ID")
		writeJsonError(res, http.StatusBadRequest, "Record ID is required")
		return
	}

	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		logger.Printf("WARN: invalid record ID, %v", err)
		writeJsonError(res, http.StatusBadRequest, "Invalid record ID")
		return
	}

	err = db.DeleteRecord(recordID)
	if err != nil {
		logger.Printf("ERROR: deleting record error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: record %d deleted successfully", recordID)
	writeJson(res, http.StatusOK, map[string]any{"message": "Record deleted successfully"})
}

func updateRecordStatusHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPut {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var statusReq UpdateStatusRequest
	if err := json.NewDecoder(req.Body).Decode(&statusReq); err != nil {
		logger.Printf("WARN: unmarshal error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}

	err := db.UpdateRecordStatus(statusReq.ID, statusReq.Status)
	if err != nil {
		logger.Printf("ERROR: updating record status error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: record %d status updated to %s", statusReq.ID, statusReq.Status)
	writeJson(res, http.StatusOK, map[string]any{"message": "Record status updated successfully"})
}

func getAllRecordsHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPost {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var pagination PaginationRequest
	if err := json.NewDecoder(req.Body).Decode(&pagination); err != nil {
		// Если пагинация не передана, используем значения по умолчанию
		pagination.Limit = 50
		pagination.Offset = 0
	}

	if pagination.Limit <= 0 || pagination.Limit > 100 {
		pagination.Limit = 50
	}
	if pagination.Offset < 0 {
		pagination.Offset = 0
	}

	records, err := db.GetAllRecords(pagination.Limit, pagination.Offset)
	if err != nil {
		logger.Printf("ERROR: getting all records error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: all records retrieved successfully (limit: %d, offset: %d)", pagination.Limit, pagination.Offset)
	writeJson(res, http.StatusOK, map[string]any{"records": records})
}

func getRecordsByStatusHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodPost {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var statusReq StatusRequest
	if err := json.NewDecoder(req.Body).Decode(&statusReq); err != nil {
		logger.Printf("WARN: unmarshal error, %v", err)
		writeJsonError(res, http.StatusBadRequest, err.Error())
		return
	}

	records, err := db.GetRecordsByStatus(statusReq.Status)
	if err != nil {
		logger.Printf("ERROR: getting records by status error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: records with status %s retrieved successfully", statusReq.Status)
	writeJson(res, http.StatusOK, map[string]any{"records": records})
}

func getRecordByIDHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	if req.Method != http.MethodGet {
		logger.Printf("WARN: incorrect request type")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	recordIDStr := req.URL.Query().Get("id")
	if recordIDStr == "" {
		logger.Printf("WARN: missing record ID")
		writeJsonError(res, http.StatusBadRequest, "Record ID is required")
		return
	}

	recordID, err := strconv.ParseInt(recordIDStr, 10, 64)
	if err != nil {
		logger.Printf("WARN: invalid record ID, %v", err)
		writeJsonError(res, http.StatusBadRequest, "Invalid record ID")
		return
	}

	record, err := db.GetRecordByID(recordID)
	if err != nil {
		logger.Printf("ERROR: getting record by ID error, %v", err)
		writeJsonError(res, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Printf("INFO: record %d retrieved successfully", recordID)
	writeJson(res, http.StatusOK, map[string]any{"record": record})
}
