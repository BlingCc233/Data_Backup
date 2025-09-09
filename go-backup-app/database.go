// database.go
package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// InitializeDatabase sets up the SQLite database in the user's app data directory.
func InitializeDatabase(ctx context.Context) (*sql.DB, error) {
	appDataDir, err := os.UserHomeDir() // 获取用户主目录
	if err != nil {
		return nil, err
	}
	dbDir := filepath.Join(appDataDir, ".gobackup")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dbDir, "history.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if it doesn't exist
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS backups (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		file_name TEXT,
		backup_path TEXT,
		created_at DATETIME
	);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}
