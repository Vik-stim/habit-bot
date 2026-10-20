package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "habits.db")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS habits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS checks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        habit_id INTEGER NOT NULL,
        date TEXT NOT NULL,
        UNIQUE(habit_id, date)
    )
`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
