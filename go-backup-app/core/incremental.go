package core

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	internalMetaPrefix = ".qbakmeta/"
	manifestEntryPath  = internalMetaPrefix + "manifest.json"
	manifestVersion    = 1
)

type BackupType string

const (
	BackupTypeFull        BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"
)

type ManifestFile struct {
	Path     string      `json:"path"`
	Size     int64       `json:"size"`
	Mode     os.FileMode `json:"mode"`
	ModTime  time.Time   `json:"modTime"`
	IsDir    bool        `json:"isDir"`
	IsLink   bool        `json:"isLink"`
	LinkDest string      `json:"linkDest,omitempty"`
}

type BackupManifest struct {
	Version   int            `json:"version"`
	Type      BackupType     `json:"type"`
	CreatedAt time.Time      `json:"createdAt"`
	Parent    string         `json:"parent,omitempty"`
	Files     []ManifestFile `json:"files"`
}

func isInternalPath(path string) bool {
	return strings.HasPrefix(path, internalMetaPrefix)
}

func (m BackupManifest) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func UnmarshalManifest(data []byte) (*BackupManifest, error) {
	var manifest BackupManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (mf ManifestFile) equalForDiff(other ManifestFile) bool {
	if mf.Path != other.Path {
		return false
	}
	if mf.IsDir != other.IsDir || mf.IsLink != other.IsLink {
		return false
	}

	// Keep mode checks lightweight and deterministic.
	if mf.Mode != other.Mode {
		return false
	}

	// For symlinks, link destination is the primary content.
	if mf.IsLink {
		return mf.LinkDest == other.LinkDest
	}

	// For dirs, ignore ModTime/Size to avoid noisy diffs caused by child changes.
	if mf.IsDir {
		return true
	}

	if mf.Size != other.Size {
		return false
	}
	return mf.ModTime.Equal(other.ModTime)
}

func manifestFilesToMap(files []ManifestFile) map[string]ManifestFile {
	m := make(map[string]ManifestFile, len(files))
	for _, f := range files {
		m[f.Path] = f
	}
	return m
}

func sortManifestFiles(files []ManifestFile) {
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
}

