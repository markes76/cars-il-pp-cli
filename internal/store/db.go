package store

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	path string
}

func DefaultPath() string {
	if value := os.Getenv("CARS_IL_DB"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "cars-il.db"
	}
	return filepath.Join(home, ".cars-il", "cars-il.db")
}

func Open(path string) (*DB, error) {
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	wrapped := &DB{DB: db, path: path}
	if err := wrapped.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return wrapped, nil
}

func (db *DB) Path() string {
	return db.path
}

func (db *DB) migrate() error {
	for _, statement := range migrationStatements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}
