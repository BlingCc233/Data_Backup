package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type archiveJob struct {
	path    string
	baseDir string
	relPath string // slash-normalized path used in archive
}

type scanResult struct {
	jobs              []archiveJob
	jobsByRelPath      map[string]archiveJob
	files             []ManifestFile
	selectedFileCount int
	selectedBytes     int64
}

func (m *BackupManager) scanSources(srcPaths []string, filters FilterConfig) (scanResult, error) {
	if len(srcPaths) == 0 {
		return scanResult{}, ErrNoFilesSelected
	}

	res := scanResult{
		jobs:         make([]archiveJob, 0, 1024),
		jobsByRelPath: make(map[string]archiveJob, 1024),
		files:        make([]ManifestFile, 0, 1024),
	}

	addEntry := func(path string, baseDir string, info os.FileInfo) error {
		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		res.jobs = append(res.jobs, archiveJob{
			path:    path,
			baseDir: baseDir,
			relPath: rel,
		})
		res.jobsByRelPath[rel] = archiveJob{
			path:    path,
			baseDir: baseDir,
			relPath: rel,
		}

		if !info.IsDir() {
			res.selectedFileCount++
			if info.Mode().IsRegular() {
				res.selectedBytes += info.Size()
			}
		}

		// Skip "." to avoid noisy diffs in incremental backups.
		if rel == "." {
			return nil
		}

		file := ManifestFile{
			Path:    rel,
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
			IsLink:  info.Mode()&os.ModeSymlink != 0,
		}
		if info.Mode().IsRegular() {
			file.Size = info.Size()
		}
		if file.IsLink {
			linkDest, err := os.Readlink(path)
			if err == nil {
				file.LinkDest = linkDest
			}
		}

		res.files = append(res.files, file)
		return nil
	}

	for _, startPath := range srcPaths {
		select {
		case <-m.ctx.Done():
			return scanResult{}, m.ctx.Err()
		default:
		}

		info, err := os.Lstat(startPath)
		if err != nil {
			return scanResult{}, fmt.Errorf("failed to stat source path %s: %w", startPath, err)
		}

		baseDir := startPath
		if !info.IsDir() {
			baseDir = filepath.Dir(startPath)
			if !filters.ShouldInclude(startPath, info) {
				continue
			}
			if err := addEntry(startPath, baseDir, info); err != nil {
				return scanResult{}, err
			}
			continue
		}

		walkErr := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			select {
			case <-m.ctx.Done():
				return context.Canceled
			default:
			}

			if !filters.ShouldInclude(path, info) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if err := addEntry(path, baseDir, info); err != nil {
				return err
			}
			return nil
		})
		if walkErr != nil {
			if errors.Is(walkErr, context.Canceled) {
				return scanResult{}, m.ctx.Err()
			}
			return scanResult{}, walkErr
		}
	}

	sortManifestFiles(res.files)
	return res, nil
}
