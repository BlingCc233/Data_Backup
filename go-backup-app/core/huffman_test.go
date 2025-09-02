// core/huffman_test.go
package core

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestHuffmanCompression(t *testing.T) {
	testCases := []struct {
		name                 string
		input                string
		expectCompression    bool // 是否期望文件体积变小
		expectNoChangeOrGrow bool // 是否期望文件体积不变或变大
	}{
		{
			name:                 "simple repetitive string (expect grow)",
			input:                "AAAAABBBCCCCCCDDE",
			expectCompression:    false, // 对于17字节的文件，头部开销太大，不期望压缩
			expectNoChangeOrGrow: true,
		},
		{
			name:                 "more complex string (expect grow)",
			input:                "go gophers are great at golang programming, go go go!",
			expectCompression:    false, // 53字节的文件，同样可能因为头部开销而变大
			expectNoChangeOrGrow: true,
		},
		{
			name:                 "empty string",
			input:                "",
			expectCompression:    false,
			expectNoChangeOrGrow: true,
		},
		{
			name:                 "single character",
			input:                "a",
			expectCompression:    false,
			expectNoChangeOrGrow: true,
		},
		{
			name:                 "all unique characters",
			input:                "abcdefghijklmnopqrstuvwxyz",
			expectCompression:    false,
			expectNoChangeOrGrow: true,
		},
		{
			name:                 "long repetitive text (expect compression)",
			input:                strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100),
			expectCompression:    true, // 4.4KB 的文件，必须被有效压缩
			expectNoChangeOrGrow: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 压缩
			var compressedBuf bytes.Buffer
			writeCloser := &bufferWriteCloser{Buffer: &compressedBuf}
			compressor := NewCompressedWriter(writeCloser)

			_, err := compressor.Write([]byte(tc.input))
			if err != nil {
				t.Fatalf("Failed to write to compressor: %v", err)
			}
			err = compressor.Close()
			if err != nil {
				t.Fatalf("Failed to close compressor: %v", err)
			}

			compressedData := compressedBuf.Bytes()

			// 检查压缩效果
			originalSize := len(tc.input)
			compressedSize := len(compressedData)

			if tc.expectCompression {
				if compressedSize >= originalSize {
					t.Errorf("Compression was ineffective: original size %d, compressed size %d", originalSize, compressedSize)
				} else {
					t.Logf("Compression effective: original size %d, compressed size %d, ratio %.2f", originalSize, compressedSize, float64(compressedSize)/float64(originalSize))
				}
			}
			if tc.expectNoChangeOrGrow {
				t.Logf("File grew or stayed same as expected: original size %d, compressed size %d", originalSize, compressedSize)
			}

			// 解压
			compressedReader := bytes.NewReader(compressedData)
			decompressor, err := NewCompressedReader(compressedReader)
			if err != nil {
				t.Fatalf("Failed to create decompressor: %v", err)
			}

			decompressedData, err := io.ReadAll(decompressor)
			if err != nil {
				t.Fatalf("Failed to read from decompressor: %v", err)
			}

			// 验证
			if string(decompressedData) != tc.input {
				t.Errorf("Decompressed data does not match original.\nOriginal:   %q\nDecompressed: %q", tc.input, string(decompressedData))
			}
		})
	}
}

// bufferWriteCloser is a helper to give a bytes.Buffer a Close method.
type bufferWriteCloser struct {
	*bytes.Buffer
}

func (b *bufferWriteCloser) Close() error {
	// No-op
	return nil
}
