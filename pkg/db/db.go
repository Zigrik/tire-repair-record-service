package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const schema string = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
	title VARCHAR NOT NULL DEFAULT "",
	comment TEXT,
	repeat VARCHAR(128)
);`

var database *sql.DB

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func CloseDatabase() {
	database.Close()
}

// we check the name and path of the database from the environment variable. If the data is incorrect, we use the default database
func checkTodoDbPath(dbFile string, logger *log.Logger) string {

	dbFileTodo := os.Getenv("TODO_DBFILE")

	if dbFileTodo == "" {
		return dbFile
	}

	if ext := strings.ToLower(filepath.Ext(dbFileTodo)); ext != ".db" {
		logger.Printf("WARN: incorrect file extension (%s), required (.db). The default database will be used.\n", ext)
		return dbFile
	}

	dir := filepath.Dir(dbFileTodo)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Printf("WARN: the %s directory does not exist. The default database will be used.\n", dir)
		return dbFile
	}

	return dbFileTodo
}

func Init(dbFile string, logger *log.Logger) error {

	dbFile = checkTodoDbPath(dbFile, logger)

	var err error
	database, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	var install bool
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		install = true
	}

	if install {
		file, err := os.Create(dbFile)
		if err != nil {
			return err
		}
		file.Close()
		logger.Printf("INFO: the %s file has been created\n", dbFile)

		_, err = database.Exec(schema)
		if err != nil {
			return err
		}
	}

	logger.Printf("INFO: the %s database is ready for use\n", dbFile)
	return nil
}
