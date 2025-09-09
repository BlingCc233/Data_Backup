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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	db     *sql.DB
	cancel context.CancelFunc // 打断备份/恢复操作

	conflictRequests map[string]*conflictRequest
	conflictMutex    sync.Mutex
	requestIDCounter int64
}

func NewApp() *App {
	return &App{
		conflictRequests: make(map[string]*conflictRequest),
	}
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

func (a *App) SelectFiles(selectDirectories bool) ([]string, error) {
	if selectDirectories {
		dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
			Title: "Select Directory",
		})
		if err != nil {
			return nil, err
		}
		if dir == "" {
			return []string{}, nil
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

// --- Feature Functions ---

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
		info, err := os.Lstat(path)
		if err != nil {
			log.Printf("Could not stat path %s: %v", path, err)
			continue
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

	if err := a.AddBackupRecord(fileName, destinationFile); err != nil {
		log.Printf("Failed to save backup record to database: %v", err)
	}

	log.Println("Backup completed successfully.")
	return "备份成功！", nil
}

// --- Restore ---

type RestoreConfig struct {
	BackupFile string `json:"backupFile"`
	RestoreDir string `json:"restoreDir"`
	Password   string `json:"password"`
}

// ResolveConflict is called by the frontend to resolve a file conflict.
func (a *App) ResolveConflict(requestID string, resolution string) error {
	a.conflictMutex.Lock()
	defer a.conflictMutex.Unlock()

	req, ok := a.conflictRequests[requestID]
	if !ok {
		return fmt.Errorf("no pending conflict request with ID: %s", requestID)
	}

	var action core.ConflictAction
	switch resolution {
	case "overwrite":
		action = core.ActionOverwrite
	case "keep_both":
		action = core.ActionKeepBoth
	case "skip":
		action = core.ActionSkip
	default:
		return fmt.Errorf("invalid resolution: %s", resolution)
	}

	req.responseChan <- action
	delete(a.conflictRequests, requestID)
	return nil
}

func (a *App) StartRestore(config RestoreConfig) (string, error) {
	opCtx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	defer func() {
		a.cancel = nil
		// 清理任何悬而未决的冲突请求
		a.conflictMutex.Lock()
		for id, req := range a.conflictRequests {
			close(req.responseChan)
			delete(a.conflictRequests, id)
		}
		a.conflictMutex.Unlock()
	}()

	log.Printf("Starting restore of %s to %s", config.BackupFile, config.RestoreDir)
	manager := core.NewBackupManager(opCtx)

	manager.ConflictHandler = func(path string) (core.ConflictAction, error) {
		a.conflictMutex.Lock()
		a.requestIDCounter++
		requestID := strconv.FormatInt(a.requestIDCounter, 10)
		req := &conflictRequest{
			responseChan: make(chan core.ConflictAction, 1),
		}
		a.conflictRequests[requestID] = req
		a.conflictMutex.Unlock()

		runtime.EventsEmit(a.ctx, "conflict_detected", map[string]string{
			"path":      path,
			"requestID": requestID,
		})

		// 等待前端的响应或操作被取消
		select {
		case <-opCtx.Done():
			return core.ActionSkip, opCtx.Err() // 如果操作被取消，默认跳过并返回错误
		case action := <-req.responseChan:
			return action, nil
		}
	}

	err := manager.Restore(config.BackupFile, config.RestoreDir, config.Password)
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
	return "恢复备份成功！", nil
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
	var validRecords []BackupRecord
	var invalidIDs []int // 存储不存在文件的记录ID

	// 先获取所有记录
	for rows.Next() {
		var r BackupRecord
		if err := rows.Scan(&r.ID, &r.FileName, &r.BackupPath, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	// 检查文件是否存在
	for _, record := range records {
		if _, err := os.Stat(record.BackupPath); err == nil {
			// 文件存在，添加到有效记录中
			validRecords = append(validRecords, record)
		} else {
			// 文件不存在，记录需要删除的ID
			invalidIDs = append(invalidIDs, record.ID)
		}
	}

	// 删除不存在的文件记录
	if len(invalidIDs) > 0 {
		// 构建占位符
		placeholders := strings.Repeat("?,", len(invalidIDs)-1) + "?"
		query := fmt.Sprintf("DELETE FROM backups WHERE id IN (%s)", placeholders)

		// 构建参数
		args := make([]interface{}, len(invalidIDs))
		for i, id := range invalidIDs {
			args[i] = id
		}

		_, err = a.db.Exec(query, args...)
		if err != nil {
			log.Printf("清理无效备份记录时出错: %v", err)
		}
	}

	return validRecords, nil
}

type conflictRequest struct {
	responseChan chan core.ConflictAction
}
