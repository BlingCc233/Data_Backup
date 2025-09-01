// core/archive_test.go
package core

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArchiveWriterAndReader(t *testing.T) {
	// 1. 准备内存 buffer 作为归档文件
	buf := new(bytes.Buffer)
	writer := NewArchiveWriter(buf)

	// 2. 准备要写入的数据和元数据
	meta1 := FileMetadata{
		Path:    "dir1/file1.txt",
		Size:    11,
		Mode:    0644,
		ModTime: time.Now(),
	}
	data1 := bytes.NewReader([]byte("hello world"))

	meta2 := FileMetadata{
		Path: "dir2",
		Size: 0,
		Mode: os.ModeDir | 0755,
	}

	// 3. 写入条目
	err := writer.WriteEntry(meta1, data1)
	assert.NoError(t, err)

	err = writer.WriteEntry(meta2, nil)
	assert.NoError(t, err)

	// 4. 创建 reader 从 buffer 读取
	reader := NewArchiveReader(buf)

	// 5. 读取并验证第一个条目
	readMeta1, err := reader.NextEntry()
	assert.NoError(t, err)
	assert.Equal(t, meta1.Path, readMeta1.Path)
	assert.Equal(t, meta1.Size, readMeta1.Size)
	assert.Equal(t, meta1.Mode, readMeta1.Mode)

	readData1 := make([]byte, readMeta1.Size)
	_, err = io.ReadFull(reader.r, readData1)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(readData1))

	// 6. 读取并验证第二个条目
	readMeta2, err := reader.NextEntry()
	assert.NoError(t, err)
	assert.Equal(t, meta2.Path, readMeta2.Path)
	assert.True(t, readMeta2.Mode.IsDir())

	// 7. 确认已经到末尾
	_, err = reader.NextEntry()
	assert.Equal(t, io.EOF, err)
}

// TODO: 添加对 filter, crypto 等模块的单元测试
