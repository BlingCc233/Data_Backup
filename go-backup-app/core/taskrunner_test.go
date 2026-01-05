package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTaskRunner_RunNowUpdatesLastBackupPath(t *testing.T) {
	calls := make(chan BackupTask, 1)
	runner := NewTaskRunner(func(ctx context.Context, task BackupTask) (string, error) {
		calls <- task
		return "/tmp/new-backup.qbak", nil
	})

	task := BackupTask{
		ID:      "t1",
		Name:    "task1",
		Type:    TaskTypeSchedule,
		Enabled: true,
		Config: TaskConfig{
			CronExpr: "@every 1h",
		},
	}
	require.NoError(t, runner.Upsert(task))

	runner.RunNow("t1")
	select {
	case <-calls:
	case <-time.After(2 * time.Second):
		t.Fatal("expected task executor to be called")
	}

	var got BackupTask
	for _, tsk := range runner.List() {
		if tsk.ID == "t1" {
			got = tsk
			break
		}
	}
	require.Equal(t, "/tmp/new-backup.qbak", got.Config.LastBackupPath)
}

func TestTaskRunner_WatchTriggersExecutor(t *testing.T) {
	tempDir := t.TempDir()

	calls := make(chan struct{}, 10)
	runner := NewTaskRunner(func(ctx context.Context, task BackupTask) (string, error) {
		calls <- struct{}{}
		return "", nil
	})
	runner.Start()
	t.Cleanup(runner.Stop)

	task := BackupTask{
		ID:      "w1",
		Name:    "watch",
		Type:    TaskTypeWatch,
		Enabled: true,
		Config: TaskConfig{
			WatchPaths:      []string{tempDir},
			WatchDebounceMs: 50,
		},
	}
	require.NoError(t, runner.Upsert(task))

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("x"), 0644))

	select {
	case <-calls:
	case <-time.After(3 * time.Second):
		t.Fatal("expected watcher to trigger task executor")
	}
}

func TestTaskRunner_ScheduleTriggersExecutor(t *testing.T) {
	calls := make(chan struct{}, 10)
	runner := NewTaskRunner(func(ctx context.Context, task BackupTask) (string, error) {
		calls <- struct{}{}
		return "", nil
	})
	runner.Start()
	t.Cleanup(runner.Stop)

	task := BackupTask{
		ID:      "s1",
		Name:    "schedule",
		Type:    TaskTypeSchedule,
		Enabled: true,
		Config: TaskConfig{
			CronExpr: "@every 1s",
		},
	}
	require.NoError(t, runner.Upsert(task))

	select {
	case <-calls:
	case <-time.After(4 * time.Second):
		t.Fatal("expected scheduled task to trigger executor")
	}
}

