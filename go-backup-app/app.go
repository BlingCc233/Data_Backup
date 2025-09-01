// app.go
package main

import (
	"context"
	"errors"
	"fmt"
	"go-backup-app/core"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectDirectory() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Directory",
	})
}

func (a *App) SelectFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Select Backup File",
		Filters: []runtime.FileFilter{{DisplayName: "GoBackup Files (*.qbak)", Pattern: "*.qbak"}},
	})
}

// BackupConfig 定义了备份任务的配置 (增加了加密字段)
type BackupConfig struct {
	SourceDir           string            `json:"sourceDir"`
	DestinationDir      string            `json:"destinationDir"`
	Filters             core.FilterConfig `json:"filters"`
	UseEncryption       bool              `json:"useEncryption"`
	EncryptionAlgorithm string            `json:"encryptionAlgorithm"` // "AES-256", "ChaCha20"
	EncryptionPassword  string            `json:"encryptionPassword"`
}

// StartBackup 暴露给前端的备份函数 (修改了逻辑以处理加密)
func (a *App) StartBackup(config BackupConfig) (string, error) {
	log.Printf("Starting backup with config: %+v\n", config)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sourceBase := filepath.Base(config.SourceDir)
	safeSourceBase := strings.ReplaceAll(sourceBase, ".", "_")
	fileName := fmt.Sprintf("%s_%s.qbak", timestamp, safeSourceBase)
	destinationFile := filepath.Join(config.DestinationDir, fileName)

	log.Printf("Generated destination file: %s", destinationFile)
	runtime.EventsEmit(a.ctx, "backup_progress", fmt.Sprintf("Backup file will be: %s", destinationFile))

	var algoID uint8
	if config.UseEncryption {
		switch config.EncryptionAlgorithm {
		case "AES-256":
			algoID = core.AlgoAES256_CTR
		case "ChaCha20":
			algoID = core.AlgoChaCha20
		default:
			return "Unsupported encryption algorithm", fmt.Errorf("unsupported algorithm: %s", config.EncryptionAlgorithm)
		}
	}

	manager := core.NewBackupManager(a.ctx)
	err := manager.Backup(
		config.SourceDir,
		destinationFile,
		config.Filters,
		config.UseEncryption,
		algoID,
		config.EncryptionPassword,
	)
	if err != nil {
		log.Printf("Backup failed: %v\n", err)
		return "Backup failed!", err
	}

	log.Println("Backup completed successfully.")
	return "Backup completed successfully!", nil
}

// StartRestore 暴露给前端的恢复函数 (增加了 password 参数)
func (a *App) StartRestore(backupFile string, restoreDir string, password string) (string, error) {
	log.Printf("Starting restore from %s to %s\n", backupFile, restoreDir)

	manager := core.NewBackupManager(a.ctx)
	err := manager.Restore(backupFile, restoreDir, password)
	if err != nil {
		// 关键：捕获需要密码的错误并通知前端
		if errors.Is(err, core.ErrPasswordRequired) {
			log.Println("Password required for restore")
			// 返回一个特殊格式的错误，前端可以解析它
			return "", fmt.Errorf("password_required")
		}
		log.Printf("Restore failed: %v\n", err)
		return "Restore failed!", err
	}

	log.Println("Restore completed successfully.")
	return "Restore completed successfully!", nil
}
