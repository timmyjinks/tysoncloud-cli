package db

import "database/sql"

func NewSqliteStorage() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:/var/lib/tysoncloud/test.db")
	if err != nil {
		return nil, err
	}

	table := `CREATE TABLE IF NOT EXISTS servers (
			id uuid PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			addr TEXT NOT NULL UNIQUE
		)`

	if _, err := db.Exec(table); err != nil {
		return nil, err
	}

	return db, nil
}
