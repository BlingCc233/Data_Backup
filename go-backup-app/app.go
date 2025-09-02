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

// BackupConfig 定义了备份任务的配置 (增加了压缩和加密字段)
type BackupConfig struct {
	SourceDir           string            `json:"sourceDir"`
	DestinationDir      string            `json:"destinationDir"`
	Filters             core.FilterConfig `json:"filters"`
	UseCompression      bool              `json:"useCompression"` // ADDED: 新增压缩选项
	UseEncryption       bool              `json:"useEncryption"`
	EncryptionAlgorithm string            `json:"encryptionAlgorithm"` // "AES-256", "ChaCha20"
	EncryptionPassword  string            `json:"encryptionPassword"`
}

// StartBackup 暴露给前端的备份函数 (修改了逻辑以处理压缩和加密)
func (a *App) StartBackup(config BackupConfig) (string, error) {
	// 现有的日志记录会自动包含新的 useCompression 字段，非常方便调试
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
	// MODIFIED: 调用 manager.Backup 时，传入新的压缩参数
	err := manager.Backup(
		config.SourceDir,
		destinationFile,
		config.Filters,
		config.UseCompression, // 传递压缩选项
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

type RestoreConfig struct {
	BackupFile string `json:"backupFile"`
	RestoreDir string `json:"restoreDir"`
	Password   string `json:"password"`
}

// StartRestore 暴露给前端的恢复函数 (无需任何改动)
// 核心逻辑中的 Restore 方法会自动检测文件是否被压缩
func (a *App) StartRestore(config RestoreConfig) (string, error) {
	log.Printf("Starting restore with config: %+v\n", config)

	manager := core.NewBackupManager(a.ctx)
	err := manager.Restore(config.BackupFile, config.RestoreDir, config.Password)
	if err != nil {
		if errors.Is(err, core.ErrPasswordRequired) {
			log.Println("Password required for restore")
			return "", fmt.Errorf("password_required")
		}
		log.Printf("Restore failed: %v\n", err)
		return "Restore failed!", err
	}

	log.Println("Restore completed successfully.")
	return "Restore completed successfully!", nil
}
