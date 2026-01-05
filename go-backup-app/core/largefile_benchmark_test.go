package core

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

const (
	benchmarkLargeFileSize = int64(128 << 20) // 128 MiB
	benchmarkPassword      = "benchmark-password"
)

func writeRepeatedFile(path string, size int64, chunk []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	remaining := size
	for remaining > 0 {
		toWrite := int64(len(chunk))
		if toWrite > remaining {
			toWrite = remaining
		}
		if _, err := f.Write(chunk[:toWrite]); err != nil {
			return err
		}
		remaining -= toWrite
	}
	return f.Sync()
}

func writePseudoRandomFile(path string, size int64, seed int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	r := rand.New(rand.NewSource(seed))
	buf := make([]byte, 1<<20) // 1 MiB

	remaining := size
	for remaining > 0 {
		if _, err := r.Read(buf); err != nil {
			return err
		}
		toWrite := int64(len(buf))
		if toWrite > remaining {
			toWrite = remaining
		}
		if _, err := f.Write(buf[:toWrite]); err != nil {
			return err
		}
		remaining -= toWrite
	}
	return f.Sync()
}

type fileGenerator struct {
	name string
	gen  func(path string) error
}

func benchmarkBackupLargeFile(b *testing.B, generator fileGenerator) {
	b.Helper()
	b.StopTimer()

	tempDir := b.TempDir()
	srcFile := filepath.Join(tempDir, "src.bin")
	if err := generator.gen(srcFile); err != nil {
		b.Fatalf("generate %s: %v", generator.name, err)
	}

	manager := NewBackupManager(context.Background())
	manager.DisableEvents()

	warmupDest := filepath.Join(tempDir, "warmup.qbak")
	if err := manager.Backup([]string{srcFile}, warmupDest, FilterConfig{MaxSize: -1}, true, true, AlgoAES256_CTR, benchmarkPassword); err != nil {
		b.Fatalf("warmup backup: %v", err)
	}
	info, err := os.Stat(warmupDest)
	if err != nil {
		b.Fatalf("stat warmup backup: %v", err)
	}
	ratio := float64(info.Size()) / float64(benchmarkLargeFileSize)
	_ = os.Remove(warmupDest)

	b.SetBytes(benchmarkLargeFileSize)
	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		dest := filepath.Join(tempDir, fmt.Sprintf("out-%d.qbak", i))
		if err := manager.Backup([]string{srcFile}, dest, FilterConfig{MaxSize: -1}, true, true, AlgoAES256_CTR, benchmarkPassword); err != nil {
			b.Fatalf("backup: %v", err)
		}
		_ = os.Remove(dest)
	}

	b.StopTimer()
	b.ReportMetric(ratio, "ratio")
}

func benchmarkRestoreLargeFile(b *testing.B, generator fileGenerator) {
	b.Helper()
	b.StopTimer()

	tempDir := b.TempDir()
	srcFile := filepath.Join(tempDir, "src.bin")
	if err := generator.gen(srcFile); err != nil {
		b.Fatalf("generate %s: %v", generator.name, err)
	}

	manager := NewBackupManager(context.Background())
	manager.DisableEvents()

	backupFile := filepath.Join(tempDir, "backup.qbak")
	if err := manager.Backup([]string{srcFile}, backupFile, FilterConfig{MaxSize: -1}, true, true, AlgoAES256_CTR, benchmarkPassword); err != nil {
		b.Fatalf("prepare backup: %v", err)
	}
	info, err := os.Stat(backupFile)
	if err != nil {
		b.Fatalf("stat prepared backup: %v", err)
	}
	ratio := float64(info.Size()) / float64(benchmarkLargeFileSize)

	b.SetBytes(benchmarkLargeFileSize)
	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		restoreDir := filepath.Join(tempDir, fmt.Sprintf("restore-%d", i))
		if err := manager.Restore(backupFile, restoreDir, benchmarkPassword); err != nil {
			b.Fatalf("restore: %v", err)
		}
		_ = os.RemoveAll(restoreDir)
	}

	b.StopTimer()
	b.ReportMetric(ratio, "ratio")
}

func BenchmarkBackupLargeFile_CompressEncrypt_AES_Compressible(b *testing.B) {
	benchmarkBackupLargeFile(b, fileGenerator{
		name: "compressible",
		gen: func(path string) error {
			return writeRepeatedFile(path, benchmarkLargeFileSize, []byte("the quick brown fox jumps over the lazy dog\n"))
		},
	})
}

func BenchmarkRestoreLargeFile_CompressEncrypt_AES_Compressible(b *testing.B) {
	benchmarkRestoreLargeFile(b, fileGenerator{
		name: "compressible",
		gen: func(path string) error {
			return writeRepeatedFile(path, benchmarkLargeFileSize, []byte("the quick brown fox jumps over the lazy dog\n"))
		},
	})
}

func BenchmarkBackupLargeFile_CompressEncrypt_AES_Random(b *testing.B) {
	benchmarkBackupLargeFile(b, fileGenerator{
		name: "random",
		gen: func(path string) error {
			return writePseudoRandomFile(path, benchmarkLargeFileSize, 1)
		},
	})
}

func BenchmarkRestoreLargeFile_CompressEncrypt_AES_Random(b *testing.B) {
	benchmarkRestoreLargeFile(b, fileGenerator{
		name: "random",
		gen: func(path string) error {
			return writePseudoRandomFile(path, benchmarkLargeFileSize, 1)
		},
	})
}
