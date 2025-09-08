// core/archive.go
package core

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// FileMetadata 存储文件的元数据
type FileMetadata struct {
	Path     string      `json:"path"`     // 相对路径
	Size     int64       `json:"size"`     // 文件大小
	Mode     os.FileMode `json:"mode"`     // 文件权限和模式
	ModTime  time.Time   `json:"modTime"`  // 修改时间
	IsDir    bool        `json:"isDir"`    // 是否是目录 (新增)
	IsLink   bool        `json:"isLink"`   // 是否是符号链接
	LinkDest string      `json:"linkDest"` // 符号链接目标
}

// ArchiveWriter 写入自定义格式的归档文件
type ArchiveWriter struct {
	w io.Writer
}

func NewArchiveWriter(w io.Writer) *ArchiveWriter {
	return &ArchiveWriter{w: w}
}

// WriteEntry 将一个文件或目录写入归档
func (aw *ArchiveWriter) WriteEntry(meta FileMetadata, data io.Reader, buffer []byte) error {
	headerBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal header: %w", err)
	}

	headerLen := uint32(len(headerBytes))
	if err := binary.Write(aw.w, binary.BigEndian, headerLen); err != nil {
		return fmt.Errorf("failed to write header length: %w", err)
	}

	if _, err := aw.w.Write(headerBytes); err != nil {
		return fmt.Errorf("failed to write header json: %w", err)
	}

	if data != nil && meta.Size > 0 {
		limitedReader := io.LimitReader(data, meta.Size)
		n, err := io.CopyBuffer(aw.w, limitedReader, buffer)
		if err != nil {
			return fmt.Errorf("failed to write file data: %w", err)
		}
		if n != meta.Size {
			return fmt.Errorf("file size mismatch for %s: expected %d, wrote %d", meta.Path, meta.Size, n)
		}
	}

	return nil
}

// ArchiveReader 读取自定义格式的归档文件
type ArchiveReader struct {
	r io.Reader
}

func NewArchiveReader(r io.Reader) *ArchiveReader {
	return &ArchiveReader{r: r}
}

// NextEntry 读取下一个文件条目。如果到文件末尾，返回 io.EOF
func (ar *ArchiveReader) NextEntry() (*FileMetadata, error) {
	var headerLen uint32
	if err := binary.Read(ar.r, binary.BigEndian, &headerLen); err != nil {
		if err == io.EOF {
			return nil, io.EOF // 正常结束
		}
		return nil, err
	}

	headerBytes := make([]byte, headerLen)
	if _, err := io.ReadFull(ar.r, headerBytes); err != nil {
		return nil, fmt.Errorf("failed to read header json: %w", err)
	}

	var meta FileMetadata
	if err := json.Unmarshal(headerBytes, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	return &meta, nil
}
