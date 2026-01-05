package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackupRestore_CRCDetectsCorruption(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("hello world"), 0644))

	destFile := filepath.Join(tempDir, "out.qbak")
	manager := NewBackupManager(context.Background())
	manager.DisableEvents()

	filters := FilterConfig{MaxSize: -1}
	require.NoError(t, manager.Backup([]string{srcDir}, destFile, filters, false, false, 0, ""))

	restoreDir := filepath.Join(tempDir, "restore")
	require.NoError(t, os.MkdirAll(restoreDir, 0755))
	require.NoError(t, manager.Restore(destFile, restoreDir, ""))
	got, err := os.ReadFile(filepath.Join(restoreDir, "a.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello world", string(got))

	// Corrupt the backup file (flip the last byte in CRC trailer).
	raw, err := os.ReadFile(destFile)
	require.NoError(t, err)
	require.Greater(t, len(raw), 0)
	raw[len(raw)-1] ^= 0xFF
	require.NoError(t, os.WriteFile(destFile, raw, 0644))

	restoreDir2 := filepath.Join(tempDir, "restore2")
	require.NoError(t, os.MkdirAll(restoreDir2, 0755))
	err = manager.Restore(destFile, restoreDir2, "")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "crc32 mismatch"), "expected crc32 mismatch error, got: %v", err)
}

