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

	// Create backups table if it doesn't exist
	sqlStmtBackups := `
	CREATE TABLE IF NOT EXISTS backups (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		file_name TEXT,
		backup_path TEXT,
		source_paths TEXT, 
		created_at DATETIME
	);
	`
	_, err = db.Exec(sqlStmtBackups)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Create profiles table if it doesn't exist
	sqlStmtProfiles := `
    CREATE TABLE IF NOT EXISTS profiles (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        paths TEXT NOT NULL
    );
    `
	_, err = db.Exec(sqlStmtProfiles)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Create tasks table if it doesn't exist
	sqlStmtTasks := `
    CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        type TEXT NOT NULL,
        enabled INTEGER NOT NULL DEFAULT 0,
        config_json TEXT NOT NULL,
        created_at DATETIME,
        updated_at DATETIME
    );
    `
	_, err = db.Exec(sqlStmtTasks)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
