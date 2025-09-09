// core/manager.go
package core

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ConflictAction defines the action to take when a file conflict occurs during restore.
type ConflictAction int

const (
	ActionSkip ConflictAction = iota
	ActionOverwrite
	ActionKeepBoth
)

// ConflictHandler is a function type that resolves file conflicts.
type ConflictHandler func(path string) ConflictAction

type BackupManager struct {
	ctx             context.Context
	ConflictHandler ConflictHandler
}

func NewBackupManager(ctx context.Context) *BackupManager {
	return &BackupManager{ctx: ctx}
}

const (
	backupWorkers  = 8
	restoreWorkers = 8
	copyBufferSize = 32 * 1024
)

func (m *BackupManager) emitLog(message string) {
	runtime.EventsEmit(m.ctx, "log_message", message)
}

func (m *BackupManager) emitProgress(message string, current, total int) {
	runtime.EventsEmit(m.ctx, "progress_update", map[string]interface{}{
		"message": message,
		"current": current,
		"total":   total,
	})
}

// Backup has been updated to accept a slice of source paths.
func (m *BackupManager) Backup(srcPaths []string, destFile string, filters FilterConfig, useCompression bool, useEncryption bool, algorithm uint8, password string) error {
	outFile, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	var writer io.WriteCloser = outFile
	if useEncryption {
		m.emitProgress("正在加密...", 0, 0)
		encryptedWriter, err := NewEncryptedWriter(writer, password, algorithm)
		if err != nil {
			return fmt.Errorf("failed to create encrypted writer: %w", err)
		}
		writer = encryptedWriter
		defer writer.Close()
	}

	if useCompression {
		m.emitProgress("正在压缩...", 0, 0)
		compressedWriter := NewCompressedWriter(writer)
		writer = compressedWriter
		defer writer.Close()
	}

	archiveWriter := NewArchiveWriter(writer)
	archiveMutex := &sync.Mutex{}

	// 计算总文件数（用于进度）
	totalFiles := 0
	for _, startPath := range srcPaths {
		info, err := os.Stat(startPath)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			totalFiles++
		} else {
			filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
				if err == nil && filters.ShouldInclude(path, info) {
					totalFiles++
				}
				return nil
			})
		}
	}

	pathsChan := make(chan string)
	errChan := make(chan error, backupWorkers)
	var wg sync.WaitGroup

	for i := 0; i < backupWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := make([]byte, copyBufferSize)
			for path := range pathsChan {
				select {
				case <-m.ctx.Done():
					return
				default:
				}

				// Determine base directory for relative path calculation
				var baseDir string
				for _, srcPath := range srcPaths {
					if strings.HasPrefix(path, srcPath) {
						// If srcPath is a file, its parent is the baseDir
						info, _ := os.Stat(srcPath)
						if info != nil && !info.IsDir() {
							baseDir = filepath.Dir(srcPath)
						} else {
							baseDir = srcPath
						}
						break
					}
				}
				if baseDir == "" {
					errChan <- fmt.Errorf("could not determine base directory for path: %s", path)
					continue
				}

				info, err := os.Lstat(path)
				if err != nil {
					errChan <- fmt.Errorf("failed to stat %s: %w", path, err)
					continue
				}

				if !filters.ShouldInclude(path, info) {
					continue
				}

				relPath, err := filepath.Rel(baseDir, path)
				if err != nil {
					errChan <- fmt.Errorf("failed to get relative path for %s: %w", path, err)
					continue
				}

				meta := FileMetadata{
					Path:    filepath.ToSlash(relPath),
					Size:    info.Size(),
					Mode:    info.Mode(),
					ModTime: info.ModTime(),
					IsDir:   info.IsDir(),
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
					defer file.Close()
					fileReader = bufio.NewReaderSize(file, copyBufferSize)
				} else {
					meta.Size = 0
				}

				m.emitLog(fmt.Sprintf("正在归档: %s", meta.Path))

				archiveMutex.Lock()
				err = archiveWriter.WriteEntry(meta, fileReader, buffer)
				archiveMutex.Unlock()

				if err != nil {
					errChan <- fmt.Errorf("failed to archive %s: %w", path, err)
				}
			}
		}()
	}

	// 多文件（夹）并行读取
	go func() {
		defer close(pathsChan)
		for _, startPath := range srcPaths {
			select {
			case <-m.ctx.Done():
				return
			default:
			}
			info, err := os.Stat(startPath)
			if err != nil {
				errChan <- err
				return
			}
			if !info.IsDir() { // 如果是文件
				pathsChan <- startPath
			} else { // 如果是文件夹
				walkErr := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					select {
					case <-m.ctx.Done():
						return context.Canceled
					case pathsChan <- path:
					}
					return nil
				})
				if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
					errChan <- walkErr
				}
			}
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	m.emitProgress("备份完成", totalFiles, totalFiles)
	return nil
}

func (m *BackupManager) getReaderPipe(backupFile string, password string) (io.ReadCloser, error) {
	inFile, err := os.Open(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}

	var reader io.Reader = inFile
	var fileCloser io.Closer = inFile

	bufReader := bufio.NewReaderSize(reader, copyBufferSize)

	magic, err := bufReader.Peek(len(magicHeader))
	if err == nil && bytes.Equal(magic, magicHeader) {
		log.Println("Encrypted file detected.")
		if password == "" {
			fileCloser.Close()
			return nil, ErrPasswordRequired
		}

		decryptedReader, err := NewDecryptedReader(bufReader, password)
		if err != nil {
			fileCloser.Close()
			return nil, fmt.Errorf("failed to create decrypted reader: %w", err)
		}
		reader = decryptedReader
	} else {
		reader = bufReader
	}

	bufReaderForCompression := bufio.NewReaderSize(reader, copyBufferSize)
	magic, err = bufReaderForCompression.Peek(len(huffmanMagic))
	if err == nil && bytes.Equal(magic, huffmanMagic) {
		log.Println("Compressed data detected.")
		compressedReader, err := NewCompressedReader(bufReaderForCompression)
		if err != nil {
			fileCloser.Close()
			return nil, fmt.Errorf("failed to create decompressor: %w", err)
		}
		reader = compressedReader
	} else {
		reader = bufReaderForCompression
	}

	return struct {
		io.Reader
		io.Closer
	}{
		Reader: reader,
		Closer: fileCloser,
	}, nil
}

// Restore - 解密并解压缩
func (m *BackupManager) Restore(backupFile, restoreDir, password string) error {
	m.emitProgress("正在准备恢复...", 0, 0)

	reader, err := m.getReaderPipe(backupFile, password)
	if err != nil {
		return err
	}
	defer reader.Close()

	archiveReader := NewArchiveReader(reader)
	return m.runRestore(archiveReader, restoreDir)
}

// runRestore 并行、分块恢复文件
func (m *BackupManager) runRestore(archiveReader *ArchiveReader, restoreDir string) error {
	m.emitProgress("正在扫描备份文件...", 0, 0)

	jobsChan := make(chan func(), restoreWorkers)
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	for i := 0; i < restoreWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-m.ctx.Done():
					return
				case task, ok := <-jobsChan:
					if !ok {
						return
					}
					task()
				}
			}
		}()
	}

	producerErr := func() error {
		defer close(jobsChan)
		buffer := make([]byte, copyBufferSize)

		for {
			select {
			case <-m.ctx.Done():
				return m.ctx.Err()
			default:
			}

			meta, err := archiveReader.NextEntry()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to read next archive entry: %w", err)
			}

			select {
			case <-m.ctx.Done():
				return m.ctx.Err()
			default:
			}

			destPath := filepath.Join(restoreDir, meta.Path)

			switch {
			case meta.IsLink, meta.IsDir:
				select {
				case <-m.ctx.Done():
					return m.ctx.Err()
				default:
				}
				jobsChan <- func() {
					select {
					case <-m.ctx.Done():
						return
					default:
					}

					err := m.createDirOrLink(meta, destPath)
					if err != nil {
						select {
						case errChan <- err:
						default:
						}
					}
				}
			case meta.Mode.IsRegular():
				select {
				case <-m.ctx.Done():
					return m.ctx.Err()
				default:
				}

				pr, pw := io.Pipe()

				jobsChan <- func() {
					select {
					case <-m.ctx.Done():
						pr.CloseWithError(m.ctx.Err())
						return
					default:
					}

					err := m.writeFileFromPipe(meta, destPath, pr, buffer)
					if err != nil {
						pr.CloseWithError(err)
						select {
						case errChan <- err:
						default:
						}
					}
				}

				limitedReader := io.LimitReader(archiveReader.r, meta.Size)
				_, err = io.Copy(pw, limitedReader)
				if err != nil {
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

	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	default:
	}

	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

func (m *BackupManager) createDirOrLink(meta *FileMetadata, destPath string) error {
	if _, err := os.Lstat(destPath); err == nil {
		// TODO
		// 文件夹冲突处理
	}

	if meta.IsLink {
		if err := os.Symlink(meta.LinkDest, destPath); err != nil {
			log.Printf("Warn: could not create symlink %s -> %s: %v", destPath, meta.LinkDest, err)
		}
	} else if meta.IsDir {
		if err := os.MkdirAll(destPath, meta.Mode.Perm()); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destPath, err)
		}
	}

	if err := os.Chmod(destPath, meta.Mode.Perm()); err != nil {
		log.Printf("Warn: could not chmod %s: %v", destPath, err)
	}
	_ = os.Chtimes(destPath, meta.ModTime, meta.ModTime)
	return nil
}

func (m *BackupManager) writeFileFromPipe(meta *FileMetadata, destPath string, pr *io.PipeReader, buffer []byte) error {
	defer pr.Close()

	// Conflict Resolution
	if _, err := os.Lstat(destPath); err == nil {
		if m.ConflictHandler != nil {
			action := m.ConflictHandler(destPath)
			switch action {
			case ActionSkip:
				m.emitLog(fmt.Sprintf("Skipping existing file: %s", destPath))
				return nil
			case ActionKeepBoth:
				// Find a new name, e.g., file.txt -> file (1).txt
				dir, file := filepath.Split(destPath)
				ext := filepath.Ext(file)
				base := strings.TrimSuffix(file, ext)
				for i := 1; ; i++ {
					newName := fmt.Sprintf("%s (%d)%s", base, i, ext)
					newPath := filepath.Join(dir, newName)
					if _, err := os.Lstat(newPath); os.IsNotExist(err) {
						destPath = newPath
						break
					}
				}
				m.emitLog(fmt.Sprintf("Keeping both, restoring to: %s", destPath))
			case ActionOverwrite:
				m.emitLog(fmt.Sprintf("Overwriting existing file: %s", destPath))
				// Proceed to create/truncate
			}
		}
	}

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

	if err := outFile.Chmod(meta.Mode.Perm()); err != nil {
		log.Printf("Warn: could not chmod %s: %v", destPath, err)
	}
	_ = os.Chtimes(destPath, meta.ModTime, meta.ModTime)
	return nil
}
