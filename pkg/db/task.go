package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

func AddTask(task *Task) (int64, error) {

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`
	res, err := database.Exec(query,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert task: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func Tasks(limit int, search, dateForm string) ([]*Task, error) {

	var rows *sql.Rows
	var err error
	var date time.Time

	date, _ = time.Parse("02.01.2006", search)
	switch date.IsZero() {
	case false:
		query := `SELECT id, date, title, comment, repeat 
					FROM scheduler 
					WHERE date = ? 
					ORDER BY date ASC, id ASC LIMIT ?`
		rows, err = database.Query(query, date.Format(dateForm), limit)
	case true:
		if search != "" {
			query := `SELECT id, date, title, comment, repeat 
						FROM scheduler 
						WHERE title LIKE CONCAT('%', ?, '%')
						OR comment LIKE CONCAT('%', ?, '%') 
						ORDER BY date ASC, id ASC LIMIT ?`
			rows, err = database.Query(query, search, search, limit)
		} else {
			query := `SELECT id, date, title, comment, repeat 
						FROM scheduler 
						ORDER BY date ASC, id ASC LIMIT ?`
			rows, err = database.Query(query, limit)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, fmt.Errorf("scan db failed: %w", err)
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while going through the records: %w", err)
	}

	if tasks == nil {
		tasks = make([]*Task, 0)
	}

	return tasks, nil
}

func GetTask(id string) (*Task, error) {

	_, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid task ID format")
	}

	query := `SELECT id, date, title, comment, repeat 
			  FROM scheduler 
			  WHERE id = ?`

	var task Task
	err = database.QueryRow(query, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

func UpdateTask(task *Task) error {

	query := `UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id `
	_, err := database.Exec(query,
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.ID),
	)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	return nil
}

func DeleteTask(id string) error {

	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `DELETE FROM scheduler WHERE id = :id`
	result, err := tx.Exec(query, sql.Named("id", id))
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("task with ID %s not found", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
