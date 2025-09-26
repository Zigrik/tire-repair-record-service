package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrTimeTooEarly   = errors.New("время записи раньше начала рабочего дня")
	ErrTimeTooLate    = errors.New("время записи позже окончания рабочего дня")
	ErrTimeNotAligned = errors.New("время записи не кратно интервалу")
	ErrTimeTooClose   = errors.New("время записи слишком близко к текущему времени")
	ErrTimeSlotTaken  = errors.New("время записи уже занято")
	ErrInvalidTime    = errors.New("некорректное время записи")
)

// ValidateRecordTime проверяет валидность времени записи
func ValidateRecordTime(db *sql.DB, recordTime time.Time) error {
	// Приводим к UTC и обнуляем секунды/наносекунды
	recordTime = recordTime.UTC().Truncate(time.Minute)
	currentTime := time.Now().UTC().Truncate(time.Minute)

	// 1. Проверка на минимальное время от текущего момента
	if recordTime.Sub(currentTime) < MinLeadTime {
		return ErrTimeTooClose
	}

	// 2. Проверка рабочего времени
	recordTimeOfDay := time.Date(0, 1, 1, recordTime.Hour(), recordTime.Minute(), 0, 0, time.UTC)

	if recordTimeOfDay.Before(StartTime) {
		return ErrTimeTooEarly
	}

	if recordTimeOfDay.After(FinishTime) || recordTimeOfDay.Equal(FinishTime) {
		return ErrTimeTooLate
	}

	// 3. Проверка кратности интервалу
	startOfDay := time.Date(recordTime.Year(), recordTime.Month(), recordTime.Day(), 0, 0, 0, 0, time.UTC)
	minutesFromStart := recordTime.Sub(startOfDay).Minutes()

	if int(minutesFromStart)%Interval != 0 {
		return ErrTimeNotAligned
	}

	// 4. Проверка занятости времени
	isTaken, err := IsTimeSlotTaken(db, recordTime)
	if err != nil {
		return fmt.Errorf("ошибка проверки занятости времени: %w", err)
	}

	if isTaken {
		return ErrTimeSlotTaken
	}

	return nil
}

// IsTimeSlotTaken проверяет, занято ли время
func IsTimeSlotTaken(db *sql.DB, recordTime time.Time) (bool, error) {
	// Рассчитываем границы интервала
	intervalStart := recordTime
	intervalEnd := recordTime.Add(time.Duration(Interval) * time.Minute)

	query := `
        SELECT COUNT(*) FROM tire_service 
        WHERE record BETWEEN ? AND ? 
        AND status = 'wait'
        AND id != ?` // исключаем текущую запись при обновлении

	var count int
	err := db.QueryRow(query, intervalStart, intervalEnd, 0).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
