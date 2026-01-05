package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackupFailsWhenNoFilesSelectedByFilters(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0755))

	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("hello"), 0644))

	destFile := filepath.Join(tempDir, "out.qbak")
	manager := NewBackupManager(context.Background())
	manager.DisableEvents()

	filters := FilterConfig{
		ExcludeNames: []string{"*.txt"},
		MaxSize:      -1,
	}

	err := manager.Backup([]string{srcDir}, destFile, filters, true, false, 0, "")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNoFilesSelected)

	_, statErr := os.Stat(destFile)
	require.True(t, os.IsNotExist(statErr), "backup file should not be created when no files are selected")
}
