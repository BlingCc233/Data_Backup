package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIncrementalBackup_RestoreChain(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("v1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "b.txt"), []byte("keep"), 0644))

	manager := NewBackupManager(context.Background())
	manager.DisableEvents()
	filters := FilterConfig{MaxSize: -1}

	baseFile := filepath.Join(tempDir, "base.qbak")
	require.NoError(t, manager.Backup([]string{srcDir}, baseFile, filters, false, false, 0, ""))

	// Apply changes: modify a, delete b, add c.
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("v2"), 0644))
	require.NoError(t, os.Remove(filepath.Join(srcDir, "b.txt")))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "c.txt"), []byte("new"), 0644))

	incFile := filepath.Join(tempDir, "inc.qbak")
	require.NoError(t, manager.BackupIncremental([]string{srcDir}, incFile, baseFile, filters, false, false, 0, ""))

	restoreDir := filepath.Join(tempDir, "restore")
	require.NoError(t, os.MkdirAll(restoreDir, 0755))
	require.NoError(t, manager.Restore(incFile, restoreDir, ""))

	gotA, err := os.ReadFile(filepath.Join(restoreDir, "a.txt"))
	require.NoError(t, err)
	require.Equal(t, "v2", string(gotA))

	gotC, err := os.ReadFile(filepath.Join(restoreDir, "c.txt"))
	require.NoError(t, err)
	require.Equal(t, "new", string(gotC))

	_, err = os.Stat(filepath.Join(restoreDir, "b.txt"))
	require.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(restoreDir, manifestEntryPath))
	require.True(t, os.IsNotExist(err), "internal manifest file should not be restored")
}

func TestIncrementalBackup_NoChanges(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("v1"), 0644))

	manager := NewBackupManager(context.Background())
	manager.DisableEvents()
	filters := FilterConfig{MaxSize: -1}

	baseFile := filepath.Join(tempDir, "base.qbak")
	require.NoError(t, manager.Backup([]string{srcDir}, baseFile, filters, false, false, 0, ""))

	incFile := filepath.Join(tempDir, "inc.qbak")
	err := manager.BackupIncremental([]string{srcDir}, incFile, baseFile, filters, false, false, 0, "")
	require.ErrorIs(t, err, ErrNoChanges)
	_, statErr := os.Stat(incFile)
	require.True(t, os.IsNotExist(statErr), "incremental file should not be created when no changes are detected")
}

