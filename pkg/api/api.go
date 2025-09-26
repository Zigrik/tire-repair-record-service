package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func writeJson(res http.ResponseWriter, status int, data interface{}) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	if err := json.NewEncoder(res).Encode(data); err != nil {
		http.Error(res, "Failed to encode response", http.StatusInternalServerError)
	}
}

func writeJsonError(res http.ResponseWriter, status int, message string) {
	writeJson(res, status, map[string]string{"error": message})
}

func taskHandler(res http.ResponseWriter, req *http.Request, logger *log.Logger) {
	defer func() {
		if err := recover(); err != nil {
			logger.Printf("WARN: panic during request processing %v", err)
			writeJsonError(res, http.StatusInternalServerError, fmt.Sprintf("%v", err))
		}
	}()
	switch req.Method {
	case http.MethodPost:
		addTaskHandler(res, req, logger)
	case http.MethodGet:
		getTaskHandler(res, req, logger)
	case http.MethodPut:
		putTaskHandler(res, req, logger)
	case http.MethodDelete:
		deleteTaskHandler(res, req, logger)
	default:
		logger.Printf("WARN: incorrect request type,")
		writeJsonError(res, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func Init(mux *http.ServeMux, logger *log.Logger) {
	mux.HandleFunc("/api/signin", func(res http.ResponseWriter, req *http.Request) { signin(res, req, logger) })

	taskHandlerWrapper := func(res http.ResponseWriter, req *http.Request) { taskHandler(res, req, logger) }
	tasksHandlerWrapper := func(res http.ResponseWriter, req *http.Request) { tasksHandler(res, req, logger) }
	taskDoneHandlerWrapper := func(res http.ResponseWriter, req *http.Request) { taskDoneHandler(res, req, logger) }

	mux.HandleFunc("/api/task", auth(taskHandlerWrapper, logger))
	mux.HandleFunc("/api/tasks", auth(tasksHandlerWrapper, logger))
	mux.HandleFunc("/api/task/done", auth(taskDoneHandlerWrapper, logger))
}

/*
GetAvailableSlots возвращает доступные временные слоты на указанную дату
AddRecordHandler обработчик добавления новой записи
UpdateRecordHandler обработчик обновления записи
DeleteRecord удаляет запись по ID
UpdateRecordStatus обновляет статус записи по ID
GetRecordsByDate возвращает все записи на определенную дату (исключая отмененные)
GetTodayRecords возвращает все записи на сегодня с возможностью фильтрации по статусу
GetRecordByID возвращает запись по ID
GetAllRecords возвращает все записи (для администрирования)
GetRecordsByStatus возвращает записи по статусу
GetPendingRecords возвращает записи в статусе ожидания
GetActiveRecords возвращает активные записи (не завершенные и не отмененные)
*/
