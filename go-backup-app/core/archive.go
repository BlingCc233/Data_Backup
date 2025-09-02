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
	IsLink   bool        `json:"isLink"`   // 是否是符号链接
	LinkDest string      `json:"linkDest"` // 符号链接目标
	// TODO: 添加 UID/GID (特定平台特定代码)
	// Uid      int         `json:"uid"`
	// Gid      int         `json:"gid"`
}

// ArchiveWriter 写入自定义格式的归档文件
type ArchiveWriter struct {
	w io.Writer
}

func NewArchiveWriter(w io.Writer) *ArchiveWriter {
	return &ArchiveWriter{w: w}
}

// WriteEntry 将一个文件或目录写入归档
func (aw *ArchiveWriter) WriteEntry(meta FileMetadata, data io.Reader) error {
	// 序列化元数据头部
	headerBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal header: %w", err)
	}

	// 写入头部长度 (4 bytes)
	headerLen := uint32(len(headerBytes))
	if err := binary.Write(aw.w, binary.BigEndian, headerLen); err != nil {
		return fmt.Errorf("failed to write header length: %w", err)
	}

	// 写入头部 JSON
	if _, err := aw.w.Write(headerBytes); err != nil {
		return fmt.Errorf("failed to write header json: %w", err)
	}

	// 写入数据长度 (8 bytes)
	dataLen := uint64(meta.Size)
	if err := binary.Write(aw.w, binary.BigEndian, dataLen); err != nil {
		return fmt.Errorf("failed to write data length: %w", err)
	}

	// 写入文件数据
	if data != nil && meta.Size > 0 {
		if _, err := io.CopyN(aw.w, data, meta.Size); err != nil {
			return fmt.Errorf("failed to write file data: %w", err)
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
		return nil, err // 如果是 EOF，则正常结束
	}

	headerBytes := make([]byte, headerLen)
	if _, err := io.ReadFull(ar.r, headerBytes); err != nil {
		return nil, fmt.Errorf("failed to read header json: %w", err)
	}

	var meta FileMetadata
	if err := json.Unmarshal(headerBytes, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	var dataLen uint64
	if err := binary.Read(ar.r, binary.BigEndian, &dataLen); err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}
	meta.Size = int64(dataLen) // 确保 meta 中的 Size 是正确的

	return &meta, nil
}
