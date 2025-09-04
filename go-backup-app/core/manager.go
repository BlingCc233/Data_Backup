// core/manager.go
package core

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type BackupManager struct {
	ctx context.Context
}

func NewBackupManager(ctx context.Context) *BackupManager {
	return &BackupManager{ctx: ctx}
}

const (
	// 定义用于备份和恢复的 worker 数量
	backupWorkers  = 8
	restoreWorkers = 8
	// 用于在复制文件时使用的缓冲区大小
	copyBufferSize = 32 * 1024
)

// Backup 执行备份
func (m *BackupManager) Backup(srcDir, destFile string, filters FilterConfig, useCompression bool, useEncryption bool, algorithm uint8, password string) error {
	// 1. 设置输出管道
	outFile, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	var writer io.WriteCloser = outFile
	if useEncryption {
		log.Println("Encryption enabled for backup.")
		encryptedWriter, err := NewEncryptedWriter(writer, password, algorithm)
		if err != nil {
			return fmt.Errorf("failed to create encrypted writer: %w", err)
		}
		writer = encryptedWriter
		defer writer.Close()
	}

	if useCompression {
		log.Println("Compression enabled for backup.")
		compressedWriter := NewCompressedWriter(writer) // 已经是并行和流式的
		writer = compressedWriter
		defer writer.Close()
	}

	archiveWriter := NewArchiveWriter(writer)
	archiveMutex := &sync.Mutex{} // ArchiveWriter 不是并发安全的，需要锁

	// 并行文件处理
	pathsChan := make(chan string)
	errChan := make(chan error, backupWorkers)
	var wg sync.WaitGroup

	// 启动 worker
	for i := 0; i < backupWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := make([]byte, copyBufferSize) // 每个 worker 有自己的缓冲区
			for path := range pathsChan {
				info, err := os.Lstat(path) // 使用 Lstat 以处理符号链接
				if err != nil {
					errChan <- fmt.Errorf("failed to stat %s: %w", path, err)
					continue
				}

				if !filters.ShouldInclude(path, info) {
					continue
				}

				relPath, err := filepath.Rel(srcDir, path)
				if err != nil {
					errChan <- fmt.Errorf("failed to get relative path for %s: %w", path, err)
					continue
				}

				meta := FileMetadata{
					Path:    filepath.ToSlash(relPath),
					Size:    info.Size(),
					Mode:    info.Mode(),
					ModTime: info.ModTime(),
				}

				var fileReader io.Reader
				if info.Mode()&os.ModeSymlink != 0 {
					linkDest, err := os.Readlink(path)
					if err != nil {
						errChan <- fmt.Errorf("failed to read link %s: %w", path, err)
						continue
					}
					meta.IsLink = true
					meta.LinkDest = linkDest
					meta.Size = 0
				} else if info.Mode().IsRegular() {
					file, err := os.Open(path)
					if err != nil {
						errChan <- fmt.Errorf("failed to open file %s: %w", path, err)
						continue
					}
					// 使用 defer file.Close() 是安全的，因为这个函数会阻塞直到文件被完全处理
					defer file.Close()
					fileReader = bufio.NewReaderSize(file, copyBufferSize)
				} else {
					meta.Size = 0 // 对于目录等
				}

				log.Printf("Backing up: %s", meta.Path)
				runtime.EventsEmit(m.ctx, "backup_progress", fmt.Sprintf("Archiving: %s", meta.Path))

				// 因为归档必须是串行的，所以这里加锁
				archiveMutex.Lock()
				err = archiveWriter.WriteEntry(meta, fileReader, buffer)
				archiveMutex.Unlock()

				if err != nil {
					errChan <- fmt.Errorf("failed to archive %s: %w", path, err)
				}
			}
		}()
	}

	// 3. 启动文件发现
	walkErr := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		pathsChan <- path
		return nil
	})

	close(pathsChan)
	wg.Wait()
	close(errChan)

	if walkErr != nil {
		return fmt.Errorf("error walking source directory: %w", walkErr)
	}

	// 检查 worker 是否有错误
	for err := range errChan {
		if err != nil {
			return err // 返回第一个遇到的错误
		}
	}

	return nil
}

// Restore 执行恢复 (并行优化)
func (m *BackupManager) Restore(backupFile, restoreDir string, password string) error {
	inFile, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer inFile.Close()

	// 统一带缓冲的读取器
	reader := bufio.NewReaderSize(inFile, copyBufferSize)

	// "偷看"流的头部判断是否为加密文件
	magic, _ := reader.Peek(len(magicHeader))
	if bytes.Equal(magic, magicHeader) {
		log.Println("Encrypted file detected, proceeding with decryption.")
		if password == "" {
			return ErrPasswordRequired
		}
		// 创建解密读取器，它会消耗掉魔术字和头部
		decryptedReader, err := NewDecryptedReader(reader, password)
		if err != nil {
			return err
		}
		// 用解密后的流重新创建一个带缓冲的读取器，为下一阶段做准备
		reader = bufio.NewReaderSize(decryptedReader, copyBufferSize)
	} else {
		log.Println("Not an encrypted file, proceeding with normal restore.")
	}

	// 处理解压
	// "偷看"当前流（可能是原始流或解密后的流）的头部
	magic, err = reader.Peek(len(huffmanMagic))
	// 如果 Peek 报错，可能是文件太短或者其他 IO 问题。
	// 如果是 EOF，说明是空文件或文件尾，不是压缩格式，是正常情况。
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to peek for compression header: %w", err)
	}
	if err == nil && bytes.Equal(magic, huffmanMagic) {
		log.Println("Compressed data detected, proceeding with decompression.")
		// 创建解压读取器，它会消耗掉魔术字
		compressedReader, err := NewCompressedReader(reader)
		if err != nil {
			return fmt.Errorf("failed to create decompressor: %w", err)
		}
		// 将 reader 替换为解压后的流
		var finalReader io.Reader = compressedReader
		archiveReader := NewArchiveReader(finalReader)
		return m.runRestore(archiveReader, restoreDir)
	}

	// 4. 如果没有压缩，直接处理归档
	log.Println("Not a compressed stream, proceeding without decompression.")
	archiveReader := NewArchiveReader(reader)
	return m.runRestore(archiveReader, restoreDir)
}

// runRestore 包含并行的恢复逻辑，从 Restore 中提取出来以提高代码清晰度
func (m *BackupManager) runRestore(archiveReader *ArchiveReader, restoreDir string) error {
	// jobsChan 现在只传递元数据和写入任务
	jobsChan := make(chan func(), restoreWorkers) // <<-- 通道类型改变
	errChan := make(chan error, 1)                // 只需要一个大小为1的错误通道
	var wg sync.WaitGroup

	// 启动 worker。worker 现在只是执行一个函数。
	for i := 0; i < restoreWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobsChan {
				task()
			}
		}()
	}

	// 启动归档读取器 (Producer)
	producerErr := func() error {
		defer close(jobsChan)
		buffer := make([]byte, copyBufferSize)

		for {
			select {
			case <-m.ctx.Done():
				return m.ctx.Err()
			default:
			}

			// 串行读取下一个条目的元数据
			meta, err := archiveReader.NextEntry()
			if err == io.EOF {
				return nil // 正常结束
			}
			if err != nil {
				return fmt.Errorf("failed to read next archive entry (archive may be corrupt): %w", err)
			}

			log.Printf("Restoring: %s", meta.Path)
			runtime.EventsEmit(m.ctx, "restore_progress", fmt.Sprintf("Extracting: %s", meta.Path))

			destPath := filepath.Join(restoreDir, meta.Path)

			// 根据条目类型，创建一个写入任务并发送给 worker
			switch {
			case meta.IsLink, meta.Mode.IsDir():
				// 对于目录和链接，创建任务很简单
				jobsChan <- func() {
					err := m.createDirOrLink(meta, destPath)
					if err != nil {
						select {
						case errChan <- err:
						default:
						}
					}
				}
			case meta.Mode.IsRegular():
				// 对于常规文件，需要流式处理
				// 创建一个管道，将读取端交给 worker
				pr, pw := io.Pipe()

				jobsChan <- func() {
					// 这是 worker 执行的任务
					err := m.writeFileFromPipe(meta, destPath, pr, buffer)
					if err != nil {
						// 如果 worker 写入失败，关闭管道以通知生产者
						pr.CloseWithError(err)
						select {
						case errChan <- err:
						default:
						}
					}
				}

				// (生产者) 串行地从归档流中读取数据，并写入管道
				//  阻塞，直到 worker 从管道中读取数据
				limitedReader := io.LimitReader(archiveReader.r, meta.Size)
				_, err = io.Copy(pw, limitedReader)

				// 关闭管道写入端，通知 worker 数据已写完
				if err != nil {
					// 如果生产者读取失败，通知 worker
					pw.CloseWithError(err)
					return fmt.Errorf("failed to copy data for %s: %w", meta.Path, err)
				}
				pw.Close()
			}
		}
	}()

	wg.Wait()
	close(errChan)

	if producerErr != nil {
		return producerErr
	}

	// 检查 worker 是否有错误
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

// --- 辅助函数 ---

// createDirOrLink 辅助函数，用于处理目录和链接的创建
func (m *BackupManager) createDirOrLink(meta *FileMetadata, destPath string) error {
	if meta.IsLink {
		if err := os.Symlink(meta.LinkDest, destPath); err != nil {
			log.Printf("Warn: could not create symlink %s -> %s: %v", destPath, meta.LinkDest, err)
		}
	} else if meta.Mode.IsDir() {
		if err := os.MkdirAll(destPath, meta.Mode); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destPath, err)
		}
	}
	// 设置权限和时间戳
	if err := os.Chmod(destPath, meta.Mode); err != nil {
		log.Printf("Warn: could not chmod %s: %v", destPath, err)
	}
	_ = os.Chtimes(destPath, meta.ModTime, meta.ModTime)
	return nil
}

// writeFileFromPipe 辅助函数，让 worker 从管道读取数据并写入文件
func (m *BackupManager) writeFileFromPipe(meta *FileMetadata, destPath string, pr *io.PipeReader, buffer []byte) error {
	defer pr.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent dir for %s: %w", destPath, err)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", destPath, err)
	}
	defer outFile.Close()

	_, err = io.CopyBuffer(outFile, pr, buffer)
	if err != nil {
		return fmt.Errorf("failed to write data to %s: %w", destPath, err)
	}

	if err := outFile.Chmod(meta.Mode); err != nil {
		log.Printf("Warn: could not chmod %s: %v", destPath, err)
	}
	if err := os.Chtimes(destPath, meta.ModTime, meta.ModTime); err != nil {
		log.Printf("Warn: could not chtimes %s: %v", destPath, err)
	}

	return nil
}
