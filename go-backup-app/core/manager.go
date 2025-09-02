// core/manager.go
package core

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log"
	"os"
	"path/filepath"
)

type BackupManager struct {
	ctx context.Context
}

func NewBackupManager(ctx context.Context) *BackupManager {
	return &BackupManager{ctx: ctx}
}

// Backup 执行备份流程 (增加了压缩和加密参数)
func (m *BackupManager) Backup(srcDir, destFile string, filters FilterConfig, useCompression bool, useEncryption bool, algorithm uint8, password string) error {
	outFile, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	var writer io.WriteCloser = outFile

	// 步骤 1: 应用加密 (如果需要)
	if useEncryption {
		log.Println("Encryption enabled for backup.")
		// NewEncryptedWriter 接收的 writer 可能是 outFile，也可能是 compressedWriter
		encryptedWriter, err := NewEncryptedWriter(writer, password, algorithm)
		if err != nil {
			return fmt.Errorf("failed to create encrypted writer: %w", err)
		}
		writer = encryptedWriter // 再次替换写入器
		defer writer.Close()
	}
	// 步骤 2: 应用压缩 (如果需要)
	// 压缩应该在加密之前，这样加密算法处理的是熵更高的数据，现在这样写就是正确的顺序
	if useCompression {
		log.Println("Compression enabled for backup.")
		compressedWriter := NewCompressedWriter(writer) // 注意：这里 writer 还是 outFile
		writer = compressedWriter                       // 将写入器替换为压缩写入器
		defer writer.Close()                            // 确保 Close 被调用以触发压缩
	}

	archiveWriter := NewArchiveWriter(writer)

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if !filters.ShouldInclude(path, info) {
			return nil
		}
		meta := FileMetadata{
			Path:    filepath.ToSlash(relPath),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
		}
		if info.Mode()&os.ModeSymlink != 0 {
			meta.IsLink = true
			linkDest, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("failed to read link %s: %w", path, err)
			}
			meta.LinkDest = linkDest
			meta.Size = 0
		}
		log.Printf("Backing up: %s", meta.Path)
		runtime.EventsEmit(m.ctx, "backup_progress", fmt.Sprintf("Archiving: %s", meta.Path))
		var fileReader io.Reader
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()
			fileReader = file
		} else {
			meta.Size = 0
		}
		return archiveWriter.WriteEntry(meta, fileReader)
	})
}

func (m *BackupManager) Restore(backupFile, restoreDir string, password string) error {
	inFile, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer inFile.Close()

	//bReader := bufio.NewReader(inFile)
	//var reader io.Reader = bReader
	var reader io.Reader = inFile

	// --- 步骤 1: 健壮地检测加密 ---
	decryptedReader, err := NewDecryptedReader(inFile, password)
	if err != nil {
		if errors.Is(err, ErrInvalidMagic) {
			// 这不是一个加密文件，是正常情况，重置文件指针并继续
			log.Println("Not an encrypted file, proceeding with normal restore.")
			_, seekErr := inFile.Seek(0, io.SeekStart)
			if seekErr != nil {
				return seekErr
			}
		} else {
			// 其他错误（包括 ErrPasswordRequired）直接返回
			return err
		}
	} else {
		// 解密成功，替换读取器
		log.Println("Encrypted file detected, proceeding with decryption.")
		reader = bufio.NewReader(decryptedReader)
	}

	// --- 步骤 2: 健壮地检测压缩 ---
	// 此时的 reader 可能是原始文件流，也可能是解密后的流
	log.Println("Compressed data detected, proceeding with decompression.")
	compressedReader, err := NewCompressedReader(reader) // NewCompressedReader 会消耗掉 magic
	if err != nil {
		return fmt.Errorf("failed to create decompressor: %w", err)
	}
	reader = compressedReader // 将 reader 替换为解压后的流
	// "偷看"当前流的头部，判断是否是压缩格式
	//huffMagicBytes, err := reader.(*bufio.Reader).Peek(len(huffmanMagic))
	//if err == nil && bytes.Equal(huffMagicBytes, huffmanMagic) {
	//	log.Println("Compressed data detected, proceeding with decompression.")
	//	compressedReader, err := NewCompressedReader(reader) // NewCompressedReader 会消耗掉 magic
	//	if err != nil {
	//		return fmt.Errorf("failed to create decompressor: %w", err)
	//	}
	//	reader = compressedReader // 将 reader 替换为解压后的流
	//} else {
	//	log.Println("Not a compressed stream, proceeding without decompression.")
	//}

	// --- 步骤 3: 读取存档 ---
	// 此时的 reader 已经是最终的、纯净的存档数据流了
	archiveReader := NewArchiveReader(reader)

	for {
		meta, err := archiveReader.NextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read next archive entry (archive may be corrupt): %w", err)
		}

		destPath := filepath.Join(restoreDir, meta.Path)
		log.Printf("Restoring: %s", destPath)
		runtime.EventsEmit(m.ctx, "restore_progress", fmt.Sprintf("Extracting: %s", meta.Path))

		// ... (恢复文件/目录/元数据的逻辑完全不变) ...
		switch {
		case meta.IsLink:
			if err := os.Symlink(meta.LinkDest, destPath); err != nil {
				log.Printf("Warn: could not create symlink %s -> %s: %v", destPath, meta.LinkDest, err)
			}
		case meta.Mode.IsDir():
			if err := os.MkdirAll(destPath, meta.Mode); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", destPath, err)
			}
		case meta.Mode.IsRegular():
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent dir for %s: %w", destPath, err)
			}
			outFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", destPath, err)
			}
			limitedReader := io.LimitReader(archiveReader.r, meta.Size)
			_, err = io.Copy(outFile, limitedReader)
			outFile.Close()
			if err != nil {
				return fmt.Errorf("failed to write data to %s: %w", destPath, err)
			}
		}
		if err := os.Chmod(destPath, meta.Mode); err != nil {
			log.Printf("Warn: could not chmod %s: %v", destPath, err)
		}
		if err := os.Chtimes(destPath, meta.ModTime, meta.ModTime); err != nil {
			log.Printf("Warn: could not chtimes %s: %v", destPath, err)
		}
	}
	return nil
}
