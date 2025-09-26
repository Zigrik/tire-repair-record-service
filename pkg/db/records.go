package db

import (
	"database/sql"
	"fmt"
	"time"
)

// GetAvailableSlots возвращает доступные временные слоты на указанную дату
func GetAvailableSlots(date time.Time) ([]time.Time, error) {
	// Нормализуем дату (начало дня)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	var availableSlots []time.Time

	// Генерируем все возможные слоты на день
	currentSlot := date.Add(StartTime.Sub(time.Time{}))
	endTime := date.Add(FinishTime.Sub(time.Time{}))

	for currentSlot.Before(endTime) {
		// Проверяем, не прошло ли время
		if currentSlot.After(time.Now().Add(MinLeadTime)) {
			// Проверяем, свободен ли слот
			isTaken, err := IsTimeSlotTaken(currentSlot)
			if err != nil {
				return nil, err
			}

			if !isTaken {
				availableSlots = append(availableSlots, currentSlot)
			}
		}

		currentSlot = currentSlot.Add(time.Duration(Interval) * time.Minute)
	}

	return availableSlots, nil
}

// AddRecord обработчик добавления новой записи
func AddRecord(record Record) error {
	// Если указано предварительное время, проверяем его
	if record.Record != nil {
		err := ValidateRecordTime(*record.Record)
		if err != nil {
			return fmt.Errorf("невалидное время записи: %w", err)
		}
	}

	// Вставляем запись в базу
	query := `
        INSERT INTO tire_service (title, record, comment, status) 
        VALUES (?, ?, ?, ?)`

	_, err := db.Exec(query, record.Title, record.Record, record.Comment, "wait")
	return err
}

// UpdateRecord обработчик обновления записи
func UpdateRecord(recordID int64, updatedRecord Record) error {
	if updatedRecord.Record != nil {
		err := ValidateRecordTime(*updatedRecord.Record)
		if err != nil {
			return fmt.Errorf("невалидное время записи: %w", err)
		}
	}

	query := `
        UPDATE tire_service 
        SET title = ?, record = ?, comment = ?, status = ?
        WHERE id = ?`

	_, err := db.Exec(query, updatedRecord.Title, updatedRecord.Record,
		updatedRecord.Comment, updatedRecord.Status, recordID)
	return err
}

// DeleteRecord удаляет запись по ID
func DeleteRecord(recordID int64) error {
	query := `DELETE FROM tire_service WHERE id = ?`

	result, err := db.Exec(query, recordID)
	if err != nil {
		return fmt.Errorf("ошибка удаления записи: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("запись с ID %d не найдена", recordID)
	}

	return nil
}

// UpdateRecordStatus обновляет статус записи по ID
func UpdateRecordStatus(recordID int64, newStatus string) error {
	// Проверяем валидность статуса
	validStatuses := map[string]bool{
		"wait":    true,
		"welcome": true,
		"in work": true,
		"done":    true,
		"cancel":  true,
	}

	if !validStatuses[newStatus] {
		return fmt.Errorf("невалидный статус: %s", newStatus)
	}

	query := `UPDATE tire_service SET status = ? WHERE id = ?`

	result, err := db.Exec(query, newStatus, recordID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("запись с ID %d не найдена", recordID)
	}

	return nil
}

// GetRecordsByDate возвращает все записи на определенную дату (исключая отмененные)
func GetRecordsByDate(date time.Time) ([]Record, error) {
	// Нормализуем дату (начало и конец дня)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
        SELECT id, date, title, record, comment, status 
        FROM tire_service 
        WHERE record BETWEEN ? AND ? 
        AND status != 'cancel'
        ORDER BY record ASC`

	rows, err := db.Query(query, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		var recordTime sql.NullTime // используем NullTime для nullable record

		err := rows.Scan(&record.ID, &record.Date, &record.Title, &recordTime, &record.Comment, &record.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи: %w", err)
		}

		if recordTime.Valid {
			record.Record = &recordTime.Time
		}

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по записям: %w", err)
	}

	return records, nil
}

// GetTodayRecords возвращает все записи на сегодня с возможностью фильтрации по статусу
func GetTodayRecords(statusFilter string) ([]Record, error) {
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var query string
	var args []interface{}

	if statusFilter == "" {
		// Без фильтра по статусу - все записи кроме отмененных
		query = `
            SELECT id, date, title, record, comment, status 
            FROM tire_service 
            WHERE record BETWEEN ? AND ? 
            AND status != 'cancel'
            ORDER BY record ASC`
		args = []interface{}{startOfDay, endOfDay}
	} else {
		// С фильтром по конкретному статусу
		query = `
            SELECT id, date, title, record, comment, status 
            FROM tire_service 
            WHERE record BETWEEN ? AND ? 
            AND status = ?
            ORDER BY record ASC`
		args = []interface{}{startOfDay, endOfDay, statusFilter}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		var recordTime sql.NullTime

		err := rows.Scan(&record.ID, &record.Date, &record.Title, &recordTime, &record.Comment, &record.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи: %w", err)
		}

		if recordTime.Valid {
			record.Record = &recordTime.Time
		}

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по записям: %w", err)
	}

	return records, nil
}

// GetRecordByID возвращает запись по ID
func GetRecordByID(recordID int64) (*Record, error) {
	query := `
        SELECT id, date, title, record, comment, status 
        FROM tire_service 
        WHERE id = ?`

	var record Record
	var recordTime sql.NullTime

	err := db.QueryRow(query, recordID).Scan(
		&record.ID, &record.Date, &record.Title, &recordTime, &record.Comment, &record.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("запись с ID %d не найдена", recordID)
		}
		return nil, fmt.Errorf("ошибка получения записи: %w", err)
	}

	if recordTime.Valid {
		record.Record = &recordTime.Time
	}

	return &record, nil
}

// GetAllRecords возвращает все записи (для администрирования)
func GetAllRecords(limit, offset int) ([]Record, error) {
	query := `
        SELECT id, date, title, record, comment, status 
        FROM tire_service 
        ORDER BY date DESC 
        LIMIT ? OFFSET ?`

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		var recordTime sql.NullTime

		err := rows.Scan(&record.ID, &record.Date, &record.Title, &recordTime, &record.Comment, &record.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи: %w", err)
		}

		if recordTime.Valid {
			record.Record = &recordTime.Time
		}

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по записям: %w", err)
	}

	return records, nil
}

// GetRecordsByStatus возвращает записи по статусу
func GetRecordsByStatus(status string) ([]Record, error) {
	query := `
        SELECT id, date, title, record, comment, status 
        FROM tire_service 
        WHERE status = ?
        ORDER BY record ASC, date ASC`

	rows, err := db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		var recordTime sql.NullTime

		err := rows.Scan(&record.ID, &record.Date, &record.Title, &recordTime, &record.Comment, &record.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи: %w", err)
		}

		if recordTime.Valid {
			record.Record = &recordTime.Time
		}

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по записям: %w", err)
	}

	return records, nil
}

// GetPendingRecords возвращает записи в статусе ожидания
func GetPendingRecords() ([]Record, error) {
	return GetRecordsByStatus("wait")
}

// GetActiveRecords возвращает активные записи (не завершенные и не отмененные)
func GetActiveRecords() ([]Record, error) {
	query := `
        SELECT id, date, title, record, comment, status 
        FROM tire_service 
        WHERE status IN ('wait', 'welcome', 'in work')
        ORDER BY record ASC, date ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		var recordTime sql.NullTime

		err := rows.Scan(&record.ID, &record.Date, &record.Title, &recordTime, &record.Comment, &record.Status)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования записи: %w", err)
		}

		if recordTime.Valid {
			record.Record = &recordTime.Time
		}

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по записям: %w", err)
	}

	return records, nil
}
