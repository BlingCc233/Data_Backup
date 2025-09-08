// app.go
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-backup-app/core"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	db     *sql.DB
	cancel context.CancelFunc // Used to stop ongoing operations
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	db, err := InitializeDatabase(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	a.db = db
}

func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		a.db.Close()
	}
}

// --- Dialogs ---

// SelectFiles is a new generic function for file/dir selection
func (a *App) SelectFiles(selectDirectories bool) ([]string, error) {
	if selectDirectories {
		// Wails doesn't have a multi-directory picker, so we use OpenDirectoryDialog and return a slice
		dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
			Title: "Select Directory",
		})
		if err != nil {
			return nil, err
		}
		if dir == "" {
			return []string{}, nil // User cancelled
		}
		return []string{dir}, nil
	}

	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Files",
	})
}

func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Directory",
	})
}

func (a *App) OpenInExplorer(path string) {
	runtime.BrowserOpenURL(a.ctx, "file://"+path)
}

// --- New Feature Functions ---

type FileInfo struct {
	Path    string    `json:"path"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    string    `json:"mode"`
	ModTime time.Time `json:"modTime"`
	IsDir   bool      `json:"isDir"`
}

func (a *App) GetFileMetadata(paths []string) ([]FileInfo, error) {
	var results []FileInfo
	for _, path := range paths {
		info, err := os.Lstat(path) // Use Lstat for symlinks
		if err != nil {
			log.Printf("Could not stat path %s: %v", path, err)
			continue // Skip files we can't access
		}
		results = append(results, FileInfo{
			Path:    path,
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
	}
	return results, nil
}

func (a *App) StopOperation() {
	if a.cancel != nil {
		log.Println("Received stop signal from frontend.")
		a.cancel()
	}
}

// --- Backup ---

type BackupConfig struct {
	SourcePaths         []string          `json:"sourcePaths"`
	DestinationDir      string            `json:"destinationDir"`
	Filters             core.FilterConfig `json:"filters"`
	UseCompression      bool              `json:"useCompression"`
	UseEncryption       bool              `json:"useEncryption"`
	EncryptionAlgorithm string            `json:"encryptionAlgorithm"`
	EncryptionPassword  string            `json:"encryptionPassword"`
}

func (a *App) StartBackup(config BackupConfig) (string, error) {
	opCtx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	defer func() { a.cancel = nil }()

	log.Printf("Starting backup with %d source paths.", len(config.SourcePaths))

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sourceBase := "backup"
	if len(config.SourcePaths) > 0 {
		sourceBase = filepath.Base(config.SourcePaths[0])
	}
	safeSourceBase := strings.ReplaceAll(sourceBase, ".", "_")
	fileName := fmt.Sprintf("%s_%s.qbak", timestamp, safeSourceBase)
	destinationFile := filepath.Join(config.DestinationDir, fileName)

	runtime.EventsEmit(a.ctx, "log_message", fmt.Sprintf("Backup file will be: %s", destinationFile))

	var algoID uint8
	if config.UseEncryption {
		switch config.EncryptionAlgorithm {
		case "AES-256":
			algoID = core.AlgoAES256_CTR
		case "ChaCha20":
			algoID = core.AlgoChaCha20
		default:
			return "", fmt.Errorf("unsupported algorithm: %s", config.EncryptionAlgorithm)
		}
	}

	manager := core.NewBackupManager(opCtx)
	err := manager.Backup(
		config.SourcePaths,
		destinationFile,
		config.Filters,
		config.UseCompression,
		config.UseEncryption,
		algoID,
		config.EncryptionPassword,
	)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Backup was cancelled by user.")
			return "Backup cancelled.", nil
		}
		log.Printf("Backup failed: %v\n", err)
		return "", fmt.Errorf("Backup failed: %w", err)
	}

	// Add to history on success
	if err := a.AddBackupRecord(fileName, destinationFile); err != nil {
		log.Printf("Failed to save backup record to database: %v", err)
	}

	log.Println("Backup completed successfully.")
	return "Backup completed successfully!", nil
}

// --- Restore ---

type RestoreConfig struct {
	BackupFile     string   `json:"backupFile"`
	RestoreDir     string   `json:"restoreDir"`
	Password       string   `json:"password"`
	FilesToRestore []string `json:"filesToRestore"`
}

func (a *App) StartRestore(config RestoreConfig) (string, error) {
	opCtx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	defer func() { a.cancel = nil }()

	log.Printf("Starting restore of %s to %s", config.BackupFile, config.RestoreDir)

	manager := core.NewBackupManager(opCtx)

	// Set up conflict handler
	manager.ConflictHandler = func(path string) core.ConflictAction {
		res, err := runtime.MessageDialog(opCtx, runtime.MessageDialogOptions{
			Type:    runtime.QuestionDialog,
			Title:   "File Conflict",
			Message: fmt.Sprintf("File already exists: %s\nWhat would you like to do?", path),
			Buttons: []string{"Overwrite", "Skip", "Keep Both"},
		})
		if err != nil {
			return core.ActionSkip // Default to skip on error
		}
		switch res {
		case "Overwrite":
			return core.ActionOverwrite
		case "Keep Both":
			return core.ActionKeepBoth
		default:
			return core.ActionSkip
		}
	}

	err := manager.Restore(config.BackupFile, config.RestoreDir, config.Password, config.FilesToRestore)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Restore was cancelled by user.")
			return "Restore cancelled.", nil
		}
		if errors.Is(err, core.ErrPasswordRequired) {
			log.Println("Password required for restore")
			return "", fmt.Errorf("password_required")
		}
		log.Printf("Restore failed: %v\n", err)
		return "", fmt.Errorf("Restore failed: %w", err)
	}

	log.Println("Restore completed successfully.")
	return "Restore completed successfully!", nil
}

// ListBackupContents reads an archive and returns its file list without extracting.
func (a *App) ListBackupContents(filePath, password string) ([]core.FileMetadata, error) {
	manager := core.NewBackupManager(a.ctx)
	contents, err := manager.ListContents(filePath, password)
	if err != nil {
		if errors.Is(err, core.ErrPasswordRequired) {
			return nil, fmt.Errorf("password_required")
		}
		return nil, err
	}
	return contents, nil
}

// --- Database Functions ---

type BackupRecord struct {
	ID         int       `json:"ID"`
	FileName   string    `json:"FileName"`
	BackupPath string    `json:"BackupPath"`
	CreatedAt  time.Time `json:"CreatedAt"`
}

func (a *App) AddBackupRecord(fileName, backupPath string) error {
	stmt, err := a.db.Prepare("INSERT INTO backups(file_name, backup_path, created_at) VALUES(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(fileName, backupPath, time.Now())
	return err
}

func (a *App) GetBackupHistory() ([]BackupRecord, error) {
	rows, err := a.db.Query("SELECT id, file_name, backup_path, created_at FROM backups ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []BackupRecord
	for rows.Next() {
		var r BackupRecord
		if err := rows.Scan(&r.ID, &r.FileName, &r.BackupPath, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}
