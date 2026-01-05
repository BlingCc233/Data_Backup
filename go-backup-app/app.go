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
	goruntime "runtime"
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

	taskRunner *core.TaskRunner
}

// Profile defines the structure for a backup profile.
type Profile struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Paths string `json:"paths"` // Newline-separated paths
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

	// Create default profiles after DB initialization
	if err := a.createDefaultProfiles(); err != nil {
		log.Printf("Warning: Could not create default profiles: %v", err)
	}

	a.initTaskRunner()
}

func (a *App) shutdown(ctx context.Context) {
	a.shutdownTaskRunner()
	if a.db != nil {
		a.db.Close()
	}
}

// createDefaultProfiles populates the database with default backup profiles for the current OS.
func (a *App) createDefaultProfiles() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get user home directory: %w", err)
	}

	profiles := make(map[string][]string)

	switch goruntime.GOOS {
	case "windows":
		profiles["文档"] = []string{filepath.Join(homeDir, "Documents")}
		profiles["照片"] = []string{filepath.Join(homeDir, "Pictures")}
	case "darwin": // macOS
		profiles["文档"] = []string{filepath.Join(homeDir, "Documents")}
		profiles["照片"] = []string{filepath.Join(homeDir, "Pictures")}
	default: // Linux and others
		profiles["文档"] = []string{filepath.Join(homeDir, "Documents")}
		profiles["照片"] = []string{filepath.Join(homeDir, "Pictures")}
	}

	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback on error

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO profiles(name, paths) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for name, paths := range profiles {
		pathsStr := strings.Join(paths, "\n")
		if _, err := stmt.Exec(name, pathsStr); err != nil {
			return fmt.Errorf("failed to insert profile '%s': %w", name, err)
		}
	}

	return tx.Commit()
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
		if errors.Is(err, core.ErrNoFilesSelected) {
			log.Println("No files selected after applying filters.")
			return "没有符合筛选条件的文件，已取消备份。", nil
		}
		log.Printf("Backup failed: %v\n", err)
		return "", fmt.Errorf("Backup failed: %w", err)
	}

	if err := a.AddBackupRecord(fileName, destinationFile, config.SourcePaths); err != nil {
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

type conflictRequest struct {
	responseChan chan core.ConflictAction
}

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
		if errors.Is(err, core.ErrInvalidPassword) {
			log.Println("Incorrect password for restore")
			return "", fmt.Errorf("password_incorrect")
		}
		log.Printf("Restore failed: %v\n", err)
		return "", fmt.Errorf("Restore failed: %w", err)
	}

	log.Println("Restore completed successfully.")
	return "恢复备份成功！", nil
}

// --- Database Functions ---

type BackupRecord struct {
	ID          int       `json:"ID"`
	FileName    string    `json:"FileName"`
	BackupPath  string    `json:"BackupPath"`
	CreatedAt   time.Time `json:"CreatedAt"`
	SourcePaths string    `json:"SourcePaths"`
}

func (a *App) AddBackupRecord(fileName, backupPath string, sourcePaths []string) error {
	// We'll store the list as a newline-separated string
	sourcePathsStr := strings.Join(sourcePaths, "\n")
	stmt, err := a.db.Prepare("INSERT INTO backups(file_name, backup_path, created_at, source_paths) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(fileName, backupPath, time.Now(), sourcePathsStr)
	return err
}

func (a *App) GetBackupHistory() ([]BackupRecord, error) {
	rows, err := a.db.Query("SELECT id, file_name, backup_path, created_at, source_paths FROM backups ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []BackupRecord
	// 使用 make 初始化为一个长度为0的非nil slice
	validRecords := make([]BackupRecord, 0)
	var invalidIDs []int

	// 先获取所有记录
	for rows.Next() {
		var r BackupRecord
		if err := rows.Scan(&r.ID, &r.FileName, &r.BackupPath, &r.CreatedAt, &r.SourcePaths); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	// 检查文件是否存在
	for _, record := range records {
		if _, err := os.Stat(record.BackupPath); err == nil {
			validRecords = append(validRecords, record)
		} else {
			invalidIDs = append(invalidIDs, record.ID)
		}
	}

	// 删除不存在的文件记录
	if len(invalidIDs) > 0 {
		placeholders := strings.Repeat("?,", len(invalidIDs)-1) + "?"
		query := fmt.Sprintf("DELETE FROM backups WHERE id IN (%s)", placeholders)

		args := make([]interface{}, len(invalidIDs))
		for i, id := range invalidIDs {
			args[i] = id
		}

		_, err = a.db.Exec(query, args...)
		if err != nil {
			log.Printf("清理无效备份记录时出错: %v", err)
		}
	}

	return validRecords, nil // 现在即使没有记录，也会返回一个 [] 而不是 nil
}

func (a *App) ListDirectory(path string) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Could not read directory %s: %v", path, err)
		return nil, fmt.Errorf("无法读取目录: %w", err)
	}

	var results []FileInfo
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info, err := os.Lstat(fullPath) // Use Lstat to handle symlinks correctly if needed
		if err != nil {
			log.Printf("Could not stat path %s: %v", fullPath, err)
			continue // Skip files we can't access
		}
		results = append(results, FileInfo{
			Path:    fullPath,
			Name:    info.Name(),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
	}
	return results, nil
}

// --- Profile Functions ---

// GetProfiles retrieves all saved backup profiles.
func (a *App) GetProfiles() ([]Profile, error) {
	rows, err := a.db.Query("SELECT id, name, paths FROM profiles ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []Profile
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.ID, &p.Name, &p.Paths); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

// CreateProfile saves a new backup profile.
func (a *App) CreateProfile(name string, paths []string) (Profile, error) {
	if name == "" {
		return Profile{}, errors.New("profile name cannot be empty")
	}
	if len(paths) == 0 {
		return Profile{}, errors.New("profile must have at least one path")
	}

	pathsStr := strings.Join(paths, "\n")

	stmt, err := a.db.Prepare("INSERT INTO profiles(name, paths) VALUES(?, ?)")
	if err != nil {
		return Profile{}, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(name, pathsStr)
	if err != nil {
		// This could be a UNIQUE constraint violation, which is a common error
		return Profile{}, fmt.Errorf("failed to execute insert for profile '%s': %w", name, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Profile{}, err
	}

	return Profile{ID: int(id), Name: name, Paths: pathsStr}, nil
}
