package core

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

func (m *BackupManager) readManifest(backupFile, password string) (*BackupManifest, error) {
	reader, err := m.getReaderPipe(backupFile, password)
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-m.ctx.Done():
			_ = reader.Close()
		case <-done:
		}
	}()
	defer close(done)
	defer reader.Close()

	archiveReader := NewArchiveReader(reader)

	meta, err := archiveReader.NextEntry()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		if m.ctx.Err() != nil {
			return nil, m.ctx.Err()
		}
		return nil, fmt.Errorf("failed to read first entry: %w", err)
	}
	if meta.Path != manifestEntryPath {
		return nil, nil
	}
	if meta.Size < 0 {
		return nil, fmt.Errorf("invalid manifest size: %d", meta.Size)
	}

	payload := make([]byte, meta.Size)
	if _, err := io.ReadFull(archiveReader.r, payload); err != nil {
		if m.ctx.Err() != nil {
			return nil, m.ctx.Err()
		}
		return nil, fmt.Errorf("failed to read manifest payload: %w", err)
	}

	// Manifest entries do not currently write CRC, but be tolerant if future versions add it.
	if meta.HasCRC {
		var ignored uint32
		if err := binary.Read(archiveReader.r, binary.BigEndian, &ignored); err != nil {
			if m.ctx.Err() != nil {
				return nil, m.ctx.Err()
			}
			return nil, fmt.Errorf("failed to read manifest crc32: %w", err)
		}
	}

	var manifest BackupManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}
	return &manifest, nil
}

func (m *BackupManager) resolveRestoreChain(backupFile, password string) ([]string, error) {
	chain := make([]string, 0, 4)
	seen := make(map[string]struct{}, 8)

	current := backupFile
	for {
		if _, ok := seen[current]; ok {
			return nil, fmt.Errorf("backup chain cycle detected at %s", current)
		}
		seen[current] = struct{}{}
		chain = append(chain, current)

		manifest, err := m.readManifest(current, password)
		if err != nil {
			return nil, err
		}
		if manifest == nil || manifest.Parent == "" || manifest.Type == BackupTypeFull {
			break
		}

		parent := manifest.Parent
		if !filepath.IsAbs(parent) {
			parent = filepath.Join(filepath.Dir(current), parent)
		}
		current = parent
	}

	// Reverse to base -> target
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain, nil
}

func (m *BackupManager) restoreSingle(backupFile, restoreDir, password string) error {
	reader, err := m.getReaderPipe(backupFile, password)
	if err != nil {
		return err
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-m.ctx.Done():
			_ = reader.Close()
		case <-done:
		}
	}()
	defer close(done)
	defer reader.Close()

	archiveReader := NewArchiveReader(reader)
	err = m.runRestore(archiveReader, restoreDir)
	if err != nil && m.ctx.Err() != nil {
		return m.ctx.Err()
	}
	return err
}

// BackupIncremental creates an incremental backup against a parent backup file.
// The parent backup must contain a manifest entry (i.e. it must be created by this version or later).
func (m *BackupManager) BackupIncremental(srcPaths []string, destFile string, parentBackupFile string, filters FilterConfig, useCompression bool, useEncryption bool, algorithm uint8, password string) error {
	if parentBackupFile == "" {
		return fmt.Errorf("parent backup file is required")
	}

	parentManifest, err := m.readManifest(parentBackupFile, password)
	if err != nil {
		return err
	}
	if parentManifest == nil {
		return fmt.Errorf("parent backup has no manifest; create a new full backup first")
	}

	scanRes, err := m.scanSources(srcPaths, filters)
	if err != nil {
		return err
	}

	currentMap := manifestFilesToMap(scanRes.files)
	parentMap := manifestFilesToMap(parentManifest.Files)

	changedSet := make(map[string]struct{}, 64)
	deletedSet := make(map[string]struct{}, 64)

	for path := range parentMap {
		if _, ok := currentMap[path]; !ok {
			deletedSet[path] = struct{}{}
		}
	}
	for path, cur := range currentMap {
		prev, ok := parentMap[path]
		if !ok {
			changedSet[path] = struct{}{}
			continue
		}
		if !cur.equalForDiff(prev) {
			changedSet[path] = struct{}{}
			if cur.IsDir != prev.IsDir || cur.IsLink != prev.IsLink {
				deletedSet[path] = struct{}{}
			}
		}
	}

	changedPaths := make([]string, 0, len(changedSet))
	for p := range changedSet {
		changedPaths = append(changedPaths, p)
	}
	sort.Strings(changedPaths)

	deletedPaths := make([]string, 0, len(deletedSet))
	for p := range deletedSet {
		deletedPaths = append(deletedPaths, p)
	}
	sort.Strings(deletedPaths)

	if len(changedPaths) == 0 && len(deletedPaths) == 0 {
		return ErrNoChanges
	}

	totalOps := len(changedPaths) + len(deletedPaths)
	var totalBytes int64
	for _, p := range changedPaths {
		if mf, ok := currentMap[p]; ok && !mf.IsDir && !mf.IsLink && mf.Size > 0 {
			totalBytes += mf.Size
		}
	}
	m.emitProgressDetail("正在归档增量...", 0, totalOps, 0, totalBytes, "archiving")

	manifest := BackupManifest{
		Version:   manifestVersion,
		Type:      BackupTypeIncremental,
		CreatedAt: time.Now(),
		Parent:    filepath.Base(parentBackupFile),
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

	var completedOps int64
	var completedBytes int64
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
			int(atomic.LoadInt64(&completedOps)),
			totalOps,
			atomic.LoadInt64(&completedBytes),
			totalBytes,
			"archiving",
		)
	}

	manifestMeta := FileMetadata{
		Path:    manifestEntryPath,
		Size:    int64(len(manifestBytes)),
		Mode:    0644,
		ModTime: time.Now(),
	}
	if err := archiveWriter.WriteEntry(manifestMeta, bytes.NewReader(manifestBytes), make([]byte, copyBufferSize), nil); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Apply deletions first to avoid conflicts when types change (file->dir, dir->file, link->file, ...).
	for _, path := range deletedPaths {
		prev := parentMap[path]
		meta := FileMetadata{
			Path:     path,
			Size:     0,
			Mode:     prev.Mode,
			ModTime:  prev.ModTime,
			IsDir:    prev.IsDir,
			IsLink:   prev.IsLink,
			LinkDest: prev.LinkDest,
			Deleted:  true,
		}
		if err := archiveWriter.WriteEntry(meta, nil, make([]byte, copyBufferSize), nil); err != nil {
			return fmt.Errorf("failed to write delete marker for %s: %w", path, err)
		}
		atomic.AddInt64(&completedOps, 1)
		emitArchivingProgress(fmt.Sprintf("正在归档: %s", meta.Path), true)
	}

	changedJobs := make([]archiveJob, 0, len(changedPaths))
	for _, p := range changedPaths {
		job, ok := scanRes.jobsByRelPath[p]
		if !ok {
			return fmt.Errorf("missing job for path %s", p)
		}
		changedJobs = append(changedJobs, job)
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
					atomic.AddInt64(&completedBytes, n)
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

				atomic.AddInt64(&completedOps, 1)
				emitArchivingProgress(fmt.Sprintf("正在归档: %s", relPath), true)
			}
		}()
	}

	go func() {
		defer close(pathsChan)
		for _, job := range changedJobs {
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

	m.emitProgressDetail("备份完成", totalOps, totalOps, totalBytes, totalBytes, "archiving")
	return nil
}

// Restore restores a backup file. If the backup is incremental, it automatically resolves and applies the chain.
func (m *BackupManager) Restore(backupFile, restoreDir, password string) error {
	m.emitProgress("正在准备恢复...", 0, 0)

	chain, err := m.resolveRestoreChain(backupFile, password)
	if err != nil {
		return err
	}

	for _, f := range chain {
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		default:
		}
		if err := m.restoreSingle(f, restoreDir, password); err != nil {
			return err
		}
	}
	return nil
}
