// core/manager.go
package core

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ConflictAction defines the action to take when a file conflict occurs during restore.
type ConflictAction int

const (
	ActionSkip ConflictAction = iota
	ActionOverwrite
	ActionKeepBoth
)

// ConflictHandler is a function type that resolves file conflicts.
type ConflictHandler func(path string) (ConflictAction, error)

type BackupManager struct {
	ctx             context.Context
	emitEvents      bool
	ConflictHandler ConflictHandler
}

func NewBackupManager(ctx context.Context) *BackupManager {
	return &BackupManager{ctx: ctx, emitEvents: true}
}

func (m *BackupManager) DisableEvents() {
	m.emitEvents = false
}

const (
	backupWorkers  = 8
	restoreWorkers = 8
	copyBufferSize = 256 * 1024
)

var restoreCopyBufferPool = sync.Pool{
	New: func() any {
		return make([]byte, copyBufferSize)
	},
}

type chainedReadCloser struct {
	r         io.Reader
	closers   []io.Closer
	closeOnce sync.Once
	closeErr  error
}

func (c *chainedReadCloser) Read(p []byte) (int, error) {
	return c.r.Read(p)
}

func (c *chainedReadCloser) Close() error {
	c.closeOnce.Do(func() {
		for i := len(c.closers) - 1; i >= 0; i-- {
			if err := c.closers[i].Close(); err != nil && c.closeErr == nil {
				c.closeErr = err
			}
		}
	})
	return c.closeErr
}

type writeCallbackWriter struct {
	w       io.Writer
	onWrite func(int)
}

func (w writeCallbackWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if n > 0 && w.onWrite != nil {
		w.onWrite(n)
	}
	return n, err
}

func (m *BackupManager) emitLog(message string) {
	if !m.emitEvents || m.ctx == nil {
		return
	}
	defer func() { _ = recover() }()
	runtime.EventsEmit(m.ctx, "log_message", message)
}

func (m *BackupManager) emitProgress(message string, current, total int) {
	m.emitProgressDetail(message, current, total, 0, 0, "")
}

func (m *BackupManager) emitProgressDetail(message string, current, total int, bytesCurrent, bytesTotal int64, stage string) {
	if !m.emitEvents || m.ctx == nil {
		return
	}
	defer func() { _ = recover() }()
	payload := map[string]interface{}{
		"message": message,
		"current": current,
		"total":   total,
	}
	if bytesTotal > 0 || bytesCurrent > 0 {
		payload["bytesCurrent"] = bytesCurrent
		payload["bytesTotal"] = bytesTotal
	}
	if stage != "" {
		payload["stage"] = stage
	}
	runtime.EventsEmit(m.ctx, "progress_update", payload)
}

// Backup has been updated to accept a slice of source paths.
func (m *BackupManager) Backup(srcPaths []string, destFile string, filters FilterConfig, useCompression bool, useEncryption bool, algorithm uint8, password string) error {
	m.emitProgressDetail("正在扫描待备份文件...", 0, 0, 0, 0, "scanning")
	scanRes, err := m.scanSources(srcPaths, filters)
	if err != nil {
		return err
	}

	if scanRes.selectedFileCount == 0 {
		return ErrNoFilesSelected
	}

	totalFiles := scanRes.selectedFileCount
	totalBytes := scanRes.selectedBytes
	m.emitProgressDetail("正在归档...", 0, totalFiles, 0, totalBytes, "archiving")

	manifest := BackupManifest{
		Version:   manifestVersion,
		Type:      BackupTypeFull,
		CreatedAt: time.Now(),
		Files:     scanRes.files,
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

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

	var archivedFiles int64
	var archivedBytes int64
	var lastProgressEmit int64
	emitArchivingProgress := func(message string, force bool) {
		now := time.Now().UnixNano()
		if !force {
			last := atomic.LoadInt64(&lastProgressEmit)
			if last != 0 && now-last < int64(150*time.Millisecond) {
				return
			}
			if !atomic.CompareAndSwapInt64(&lastProgressEmit, last, now) {
				return
			}
		} else {
			atomic.StoreInt64(&lastProgressEmit, now)
		}

		m.emitProgressDetail(
			message,
			int(atomic.LoadInt64(&archivedFiles)),
			totalFiles,
			atomic.LoadInt64(&archivedBytes),
			totalBytes,
			"archiving",
		)
	}

	// Write manifest first.
	manifestMeta := FileMetadata{
		Path:    manifestEntryPath,
		Size:    int64(len(manifestBytes)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := archiveWriter.WriteEntry(manifestMeta, bytes.NewReader(manifestBytes), make([]byte, copyBufferSize), nil); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	pathsChan := make(chan archiveJob)
	errChan := make(chan error, backupWorkers)
	var wg sync.WaitGroup

	for i := 0; i < backupWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buffer := make([]byte, copyBufferSize)
			for job := range pathsChan {
				select {
				case <-m.ctx.Done():
					return
				default:
				}

				info, err := os.Lstat(job.path)
				if err != nil {
					errChan <- fmt.Errorf("failed to stat %s: %w", job.path, err)
					continue
				}

				meta := FileMetadata{
					Path:    job.relPath,
					Size:    info.Size(),
					Mode:    info.Mode(),
					ModTime: info.ModTime(),
					IsDir:   info.IsDir(),
				}

				var fileReader io.Reader
				var openedFile *os.File
				if info.Mode()&os.ModeSymlink != 0 {
					linkDest, err := os.Readlink(job.path)
					if err != nil {
						errChan <- fmt.Errorf("failed to read link %s: %w", job.path, err)
						continue
					}
					meta.IsLink = true
					meta.LinkDest = linkDest
					meta.Size = 0
				} else if info.Mode().IsRegular() {
					meta.HasCRC = true
					file, err := os.Open(job.path)
					if err != nil {
						errChan <- fmt.Errorf("failed to open file %s: %w", job.path, err)
						continue
					}
					openedFile = file
					fileReader = bufio.NewReaderSize(file, copyBufferSize)
				} else {
					meta.Size = 0
				}

				m.emitLog(fmt.Sprintf("正在归档: %s", meta.Path))

				relPath := meta.Path
				onWrite := func(n int64) {
					atomic.AddInt64(&archivedBytes, n)
					emitArchivingProgress(fmt.Sprintf("正在归档: %s", relPath), false)
				}

				archiveMutex.Lock()
				err = archiveWriter.WriteEntry(meta, fileReader, buffer, onWrite)
				archiveMutex.Unlock()

				if openedFile != nil {
					_ = openedFile.Close()
				}

				if err != nil {
					errChan <- fmt.Errorf("failed to archive %s: %w", job.path, err)
					continue
				}

				if !meta.IsDir {
					atomic.AddInt64(&archivedFiles, 1)
					emitArchivingProgress(fmt.Sprintf("正在归档: %s", relPath), true)
				}
			}
		}()
	}

	go func() {
		defer close(pathsChan)
		for _, job := range scanRes.jobs {
			select {
			case <-m.ctx.Done():
				return
			case pathsChan <- job:
			}
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		return err
	}

	m.emitProgressDetail("备份完成", totalFiles, totalFiles, totalBytes, totalBytes, "archiving")
	return nil
}

func (m *BackupManager) getReaderPipe(backupFile string, password string) (io.ReadCloser, error) {
	inFile, err := os.Open(backupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}

	var reader io.Reader = inFile
	closers := []io.Closer{inFile}
	closeAll := func() {
		for i := len(closers) - 1; i >= 0; i-- {
			_ = closers[i].Close()
		}
	}

	bufReader := bufio.NewReaderSize(reader, copyBufferSize)

	magic, err := bufReader.Peek(len(magicHeader))
	encrypted := false
	if err == nil && bytes.Equal(magic, magicHeader) {
		log.Println("Encrypted file detected.")
		if password == "" {
			closeAll()
			return nil, ErrPasswordRequired
		}

		decryptedReader, err := NewDecryptedReader(bufReader, password)
		if err != nil {
			closeAll()
			return nil, fmt.Errorf("failed to create decrypted reader: %w", err)
		}
		encrypted = true
		closers = append(closers, decryptedReader)
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
			closeAll()
			return nil, fmt.Errorf("failed to create decompressor: %w", err)
		}
		closers = append(closers, compressedReader)
		reader = compressedReader
	} else {
		// For encrypted files, validate the next layer early so incorrect passwords fail fast.
		if encrypted {
			peek, peekErr := bufReaderForCompression.Peek(5)
			if peekErr != nil {
				closeAll()
				return nil, ErrInvalidPassword
			}
			headerLen := binary.BigEndian.Uint32(peek[:4])
			if headerLen == 0 || headerLen > maxArchiveHeaderLen || peek[4] != '{' {
				closeAll()
				return nil, ErrInvalidPassword
			}
		}
		reader = bufReaderForCompression
	}

	return &chainedReadCloser{r: reader, closers: closers}, nil
}

// runRestore 并行、分块恢复文件
func (m *BackupManager) runRestore(archiveReader *ArchiveReader, restoreDir string) error {
	m.emitProgressDetail("正在扫描备份文件...", 0, 0, 0, 0, "scanning")

	var totalFiles int64
	var totalBytes int64
	var restoredFiles int64
	var restoredBytes int64
	var lastProgressEmit int64
	emitRestoreProgress := func(message string, force bool) {
		now := time.Now().UnixNano()
		if !force {
			last := atomic.LoadInt64(&lastProgressEmit)
			if last != 0 && now-last < int64(150*time.Millisecond) {
				return
			}
			if !atomic.CompareAndSwapInt64(&lastProgressEmit, last, now) {
				return
			}
		} else {
			atomic.StoreInt64(&lastProgressEmit, now)
		}

		m.emitProgressDetail(
			message,
			int(atomic.LoadInt64(&restoredFiles)),
			int(atomic.LoadInt64(&totalFiles)),
			atomic.LoadInt64(&restoredBytes),
			atomic.LoadInt64(&totalBytes),
			"restoring",
		)
	}

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
		producerBuffer := make([]byte, copyBufferSize)

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

			// Internal metadata entries (e.g. manifest).
			if isInternalPath(meta.Path) {
				if meta.Path == manifestEntryPath {
					if meta.Size < 0 {
						return fmt.Errorf("invalid manifest size: %d", meta.Size)
					}

					payload := make([]byte, meta.Size)
					if meta.Size > 0 {
						if _, err := io.ReadFull(archiveReader.r, payload); err != nil {
							return fmt.Errorf("failed to read manifest payload: %w", err)
						}
					}
					if meta.HasCRC {
						var ignored uint32
						if err := binary.Read(archiveReader.r, binary.BigEndian, &ignored); err != nil {
							return fmt.Errorf("failed to read manifest crc32: %w", err)
						}
					}

					manifest, err := UnmarshalManifest(payload)
					if err != nil {
						return fmt.Errorf("failed to parse manifest: %w", err)
					}

					if manifest != nil && manifest.Type == BackupTypeFull {
						var files int64
						var bytes int64
						for _, f := range manifest.Files {
							if f.IsDir {
								continue
							}
							if f.IsLink || f.Mode.IsRegular() {
								files++
							}
							if !f.IsLink && f.Mode.IsRegular() && f.Size > 0 {
								bytes += f.Size
							}
						}
						atomic.StoreInt64(&totalFiles, files)
						atomic.StoreInt64(&totalBytes, bytes)
					} else {
						atomic.StoreInt64(&totalFiles, 0)
						atomic.StoreInt64(&totalBytes, 0)
					}

					emitRestoreProgress("正在恢复...", true)
					continue
				}

				if meta.Size > 0 {
					if _, err := io.CopyN(io.Discard, archiveReader.r, meta.Size); err != nil {
						return fmt.Errorf("failed to skip internal entry %s: %w", meta.Path, err)
					}
				}
				if meta.HasCRC {
					var ignored uint32
					if err := binary.Read(archiveReader.r, binary.BigEndian, &ignored); err != nil {
						return fmt.Errorf("failed to skip internal crc32 for %s: %w", meta.Path, err)
					}
				}
				continue
			}

			// Deletion marker (incremental backups).
			if meta.Deleted {
				if meta.Size > 0 {
					if _, err := io.CopyN(io.Discard, archiveReader.r, meta.Size); err != nil {
						return fmt.Errorf("failed to skip deleted entry payload for %s: %w", meta.Path, err)
					}
				}
				if meta.HasCRC {
					var ignored uint32
					if err := binary.Read(archiveReader.r, binary.BigEndian, &ignored); err != nil {
						return fmt.Errorf("failed to skip deleted crc32 for %s: %w", meta.Path, err)
					}
				}

				jobsChan <- func() {
					select {
					case <-m.ctx.Done():
						return
					default:
					}

					if err := os.RemoveAll(destPath); err != nil && !os.IsNotExist(err) {
						select {
						case errChan <- fmt.Errorf("failed to remove %s: %w", destPath, err):
						default:
						}
					}
				}
				continue
			}

			switch {
			case meta.IsLink, meta.IsDir:
				metaCopy := *meta
				destPathCopy := destPath
				relPath := metaCopy.Path
				isDir := metaCopy.IsDir
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

					err := m.createDirOrLink(&metaCopy, destPathCopy)
					if err != nil {
						select {
						case errChan <- err:
						default:
						}
						return
					}
					if !isDir {
						atomic.AddInt64(&restoredFiles, 1)
						emitRestoreProgress(fmt.Sprintf("已恢复: %s", relPath), true)
					}
				}
			case meta.Mode.IsRegular():
				select {
				case <-m.ctx.Done():
					return m.ctx.Err()
				default:
				}

				metaCopy := *meta
				destPathCopy := destPath
				relPath := metaCopy.Path
				pr, pw := io.Pipe()

				emitRestoreProgress(fmt.Sprintf("正在恢复: %s", relPath), true)

				jobsChan <- func() {
					select {
					case <-m.ctx.Done():
						pr.CloseWithError(m.ctx.Err())
						return
					default:
					}

					buffer := restoreCopyBufferPool.Get().([]byte)
					defer restoreCopyBufferPool.Put(buffer)
					err := m.writeFileFromPipe(&metaCopy, destPathCopy, pr, buffer)
					if err != nil {
						pr.CloseWithError(err)
						select {
						case errChan <- err:
							default:
						}
					}
				}

				writer := writeCallbackWriter{
					w: pw,
					onWrite: func(n int) {
						atomic.AddInt64(&restoredBytes, int64(n))
						emitRestoreProgress(fmt.Sprintf("正在恢复: %s", relPath), false)
					},
				}

				limitedReader := io.LimitReader(archiveReader.r, metaCopy.Size)
				var crcSum uint32
				if metaCopy.HasCRC {
					h := crc32.NewIEEE()
					_, err = io.CopyBuffer(writer, io.TeeReader(limitedReader, h), producerBuffer)
					crcSum = h.Sum32()
				} else {
					_, err = io.CopyBuffer(writer, limitedReader, producerBuffer)
				}
				if err != nil {
					pw.CloseWithError(err)
					return fmt.Errorf("failed to copy data for %s: %w", relPath, err)
				}

				if metaCopy.HasCRC {
					var expected uint32
					if err := binary.Read(archiveReader.r, binary.BigEndian, &expected); err != nil {
						pw.CloseWithError(err)
						return fmt.Errorf("failed to read crc32 for %s: %w", relPath, err)
					}
					if crcSum != expected {
						mismatchErr := fmt.Errorf("crc32 mismatch for %s", relPath)
						pw.CloseWithError(mismatchErr)
						return mismatchErr
					}
				}

				pw.Close()
				atomic.AddInt64(&restoredFiles, 1)
				emitRestoreProgress(fmt.Sprintf("已恢复: %s", relPath), true)
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

	emitRestoreProgress("恢复完成", true)
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
			action, err := m.ConflictHandler(destPath)
			if err != nil {
				return err
			}
			switch action {
			case ActionSkip:
				m.emitLog(fmt.Sprintf("Skipping existing file: %s", destPath))
				// BUG FIX: 必须消费掉管道中的数据，否则写入端会阻塞然后报错"write on closed pipe"
				_, _ = io.Copy(io.Discard, pr)
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
		// 检查错误是否是由于管道关闭引起的，这通常是正常情况，因为生产者完成了写入。
		if !errors.Is(err, io.ErrClosedPipe) {
			return fmt.Errorf("failed to write data to %s: %w", destPath, err)
		}
	}

	if err := outFile.Chmod(meta.Mode.Perm()); err != nil {
		log.Printf("Warn: could not chmod %s: %v", destPath, err)
	}
	_ = os.Chtimes(destPath, meta.ModTime, meta.ModTime)
	return nil
}
