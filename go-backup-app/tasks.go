package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-backup-app/core"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (a *App) initTaskRunner() {
	a.taskRunner = core.NewTaskRunner(func(ctx context.Context, task core.BackupTask) (string, error) {
		return a.executeTask(ctx, task)
	})

	tasks, err := a.loadTasksFromDB()
	if err != nil {
		log.Printf("Warning: could not load tasks: %v", err)
		return
	}
	for _, task := range tasks {
		if err := a.taskRunner.Upsert(task); err != nil {
			log.Printf("Warning: could not register task %s: %v", task.ID, err)
		}
	}

	a.taskRunner.Start()
}

func (a *App) shutdownTaskRunner() {
	if a.taskRunner != nil {
		a.taskRunner.Stop()
		a.taskRunner = nil
	}
}

func (a *App) loadTasksFromDB() ([]core.BackupTask, error) {
	if a.db == nil {
		return nil, errors.New("database not initialized")
	}

	rows, err := a.db.Query("SELECT id, name, type, enabled, config_json FROM tasks ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]core.BackupTask, 0)
	for rows.Next() {
		var (
			id        int
			name      string
			typ       string
			enabled   int
			configRaw string
		)
		if err := rows.Scan(&id, &name, &typ, &enabled, &configRaw); err != nil {
			return nil, err
		}

		var cfg core.TaskConfig
		if err := json.Unmarshal([]byte(configRaw), &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse task %d config: %w", id, err)
		}

		tasks = append(tasks, core.BackupTask{
			ID:      strconv.Itoa(id),
			Name:    name,
			Type:    core.TaskType(typ),
			Enabled: enabled != 0,
			Config:  cfg,
		})
	}
	return tasks, nil
}

func (a *App) GetTasks() ([]core.BackupTask, error) {
	return a.loadTasksFromDB()
}

func (a *App) CreateTask(task core.BackupTask) (core.BackupTask, error) {
	if a.db == nil {
		return core.BackupTask{}, errors.New("database not initialized")
	}
	if strings.TrimSpace(task.Name) == "" {
		return core.BackupTask{}, errors.New("task name cannot be empty")
	}
	if task.Type != core.TaskTypeSchedule && task.Type != core.TaskTypeWatch {
		return core.BackupTask{}, fmt.Errorf("invalid task type: %s", task.Type)
	}

	now := time.Now()
	task.Config.CreatedAt = now
	task.Config.UpdatedAt = now

	cfgBytes, err := json.Marshal(task.Config)
	if err != nil {
		return core.BackupTask{}, err
	}

	res, err := a.db.Exec(
		"INSERT INTO tasks(name, type, enabled, config_json, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?)",
		task.Name,
		string(task.Type),
		boolToInt(task.Enabled),
		string(cfgBytes),
		now,
		now,
	)
	if err != nil {
		return core.BackupTask{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return core.BackupTask{}, err
	}
	task.ID = strconv.FormatInt(id, 10)

	if a.taskRunner != nil {
		if err := a.taskRunner.Upsert(task); err != nil {
			_, _ = a.db.Exec("DELETE FROM tasks WHERE id = ?", id)
			return core.BackupTask{}, err
		}
	}

	return task, nil
}

func (a *App) UpdateTask(task core.BackupTask) error {
	if a.db == nil {
		return errors.New("database not initialized")
	}
	taskID, err := strconv.Atoi(task.ID)
	if err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}

	task.Config.UpdatedAt = time.Now()
	cfgBytes, err := json.Marshal(task.Config)
	if err != nil {
		return err
	}

	_, err = a.db.Exec(
		"UPDATE tasks SET name = ?, type = ?, enabled = ?, config_json = ?, updated_at = ? WHERE id = ?",
		task.Name,
		string(task.Type),
		boolToInt(task.Enabled),
		string(cfgBytes),
		task.Config.UpdatedAt,
		taskID,
	)
	if err != nil {
		return err
	}

	if a.taskRunner != nil {
		return a.taskRunner.Upsert(task)
	}
	return nil
}

func (a *App) DeleteTask(taskID string) error {
	if a.db == nil {
		return errors.New("database not initialized")
	}
	id, err := strconv.Atoi(taskID)
	if err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}

	_, err = a.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}

	if a.taskRunner != nil {
		a.taskRunner.Remove(taskID)
	}
	return nil
}

func (a *App) RunTaskNow(taskID string) error {
	if a.taskRunner == nil {
		return errors.New("task runner not initialized")
	}
	a.taskRunner.RunNow(taskID)
	return nil
}

func (a *App) executeTask(ctx context.Context, task core.BackupTask) (string, error) {
	if strings.TrimSpace(task.Config.DestinationDir) == "" {
		return "", errors.New("destinationDir is required")
	}
	if len(task.Config.SourcePaths) == 0 {
		return "", errors.New("sourcePaths is required")
	}

	if err := os.MkdirAll(task.Config.DestinationDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	name := strings.TrimSpace(task.Name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, string(os.PathSeparator), "_")
	fileName := fmt.Sprintf("%s_%s.qbak", timestamp, name)
	destinationFile := filepath.Join(task.Config.DestinationDir, fileName)

	manager := core.NewBackupManager(ctx)
	manager.DisableEvents()

	var backupErr error
	if task.Config.Incremental && task.Config.LastBackupPath != "" {
		backupErr = manager.BackupIncremental(
			task.Config.SourcePaths,
			destinationFile,
			task.Config.LastBackupPath,
			task.Config.Filters,
			task.Config.UseCompression,
			task.Config.UseEncryption,
			task.Config.Algorithm,
			task.Config.Password,
		)
		if errors.Is(backupErr, core.ErrNoChanges) {
			return "", nil
		}
	} else {
		backupErr = manager.Backup(
			task.Config.SourcePaths,
			destinationFile,
			task.Config.Filters,
			task.Config.UseCompression,
			task.Config.UseEncryption,
			task.Config.Algorithm,
			task.Config.Password,
		)
	}
	if backupErr != nil {
		return "", backupErr
	}

	// Save record to history (best-effort).
	if err := a.AddBackupRecord(fileName, destinationFile, task.Config.SourcePaths); err != nil {
		log.Printf("Failed to save backup record to database: %v", err)
	}

	// Persist last backup path back to task config (best-effort).
	task.Config.LastBackupPath = destinationFile
	if err := a.updateTaskConfig(task.ID, task.Config); err != nil {
		log.Printf("Failed to update task %s last backup path: %v", task.ID, err)
	}

	return destinationFile, nil
}

func (a *App) updateTaskConfig(taskID string, cfg core.TaskConfig) error {
	if a.db == nil {
		return errors.New("database not initialized")
	}
	id, err := strconv.Atoi(taskID)
	if err != nil {
		return fmt.Errorf("invalid task id: %w", err)
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	_, err = a.db.Exec(
		"UPDATE tasks SET config_json = ?, updated_at = ? WHERE id = ?",
		string(cfgBytes),
		time.Now(),
		id,
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
