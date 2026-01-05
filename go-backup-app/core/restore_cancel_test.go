package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRestoreWrongPasswordFailsFast(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "src.bin")
	require.NoError(t, os.WriteFile(srcFile, []byte("hello world"), 0644))

	backupFile := filepath.Join(tempDir, "backup.qbak")
	manager := NewBackupManager(context.Background())
	manager.DisableEvents()

	const correct = "correct-password"
	require.NoError(t, manager.Backup(
		[]string{srcFile},
		backupFile,
		FilterConfig{MaxSize: -1},
		true,
		true,
		AlgoAES256_CTR,
		correct,
	))

	start := time.Now()
	err := manager.Restore(backupFile, filepath.Join(tempDir, "restore"), "wrong-password")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidPassword)
	require.Less(t, time.Since(start), 2*time.Second)
}

func TestRestoreCancelStopsQuickly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large restore cancel test in -short")
	}

	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "big.bin")

	const fileSize = int64(32 << 20) // 32 MiB
	require.NoError(t, writePseudoRandomFile(srcFile, fileSize, 1))

	backupFile := filepath.Join(tempDir, "backup.qbak")

	ctx, cancel := context.WithCancel(context.Background())
	manager := NewBackupManager(ctx)
	manager.DisableEvents()

	const password = "pw"
	require.NoError(t, manager.Backup(
		[]string{srcFile},
		backupFile,
		FilterConfig{MaxSize: -1},
		true,
		true,
		AlgoAES256_CTR,
		password,
	))

	done := make(chan error, 1)
	restoreDir := filepath.Join(tempDir, "restore")
	go func() {
		done <- manager.Restore(backupFile, restoreDir, password)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("restore did not stop after cancel")
	}
}

