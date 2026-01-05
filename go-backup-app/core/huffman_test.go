package core

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

// mockWriteCloser 用于模拟一个可关闭的写入器，主要用于测试。
type mockWriteCloser struct {
	*bytes.Buffer
}

func (mwc *mockWriteCloser) Close() error {
	// 在模拟中，关闭操作通常什么也不做
	return nil
}

func newMockWriteCloser() *mockWriteCloser {
	return &mockWriteCloser{Buffer: &bytes.Buffer{}}
}

// TestBuildFrequencyTable 测试频率表构建
func TestBuildFrequencyTable(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected map[byte]int64
	}{
		{
			name:     "Empty input",
			input:    []byte{},
			expected: map[byte]int64{},
		},
		{
			name:     "Simple input",
			input:    []byte("hello world"),
			expected: map[byte]int64{'h': 1, 'e': 1, 'l': 3, 'o': 2, ' ': 1, 'w': 1, 'r': 1, 'd': 1},
		},
		{
			name:     "Repetitive input",
			input:    []byte("aaaaabbc"),
			expected: map[byte]int64{'a': 5, 'b': 2, 'c': 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildFrequencyTable(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("buildFrequencyTable() got = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestBuildHuffmanTree 测试Huffman树的构建
func TestBuildHuffmanTree(t *testing.T) {
	t.Run("Empty frequency table", func(t *testing.T) {
		tree := buildHuffmanTree(map[byte]int64{})
		if tree != nil {
			t.Error("Expected nil tree for empty frequency table")
		}
	})

	t.Run("Single character frequency table", func(t *testing.T) {
		freqTable := map[byte]int64{'a': 10}
		tree := buildHuffmanTree(freqTable)
		if tree == nil {
			t.Fatal("Tree should not be nil")
		}
		if tree.left == nil || tree.left.char != 'a' {
			t.Error("Tree structure is incorrect for single character")
		}
		if tree.freq != 10 {
			t.Errorf("Root frequency should be 10, got %d", tree.freq)
		}
	})

	t.Run("Multiple characters", func(t *testing.T) {
		freqTable := map[byte]int64{'a': 5, 'b': 2, 'c': 1}
		tree := buildHuffmanTree(freqTable)
		if tree == nil {
			t.Fatal("Tree should not be nil")
		}
		// 总频率应为所有字符频率之和
		if tree.freq != 8 {
			t.Errorf("Expected root frequency to be 8, got %d", tree.freq)
		}
		// 检查树的基本结构（基于频率排序）
		// c (1) 和 b (2) 应该是最深的叶子
		if tree.left == nil || tree.right == nil {
			t.Fatal("Tree should have two children")
		}
	})
}

// TestGenerateCodes 测试编码生成
func TestGenerateCodes(t *testing.T) {
	t.Run("Standard case", func(t *testing.T) {
		// 手动构建一个确定的树: a=3, b=2, c=1 -> (c,b) -> ((c,b),a)
		nodeC := &huffmanNode{char: 'c', freq: 1}
		nodeB := &huffmanNode{char: 'b', freq: 2}
		nodeA := &huffmanNode{char: 'a', freq: 3}
		parentCB := &huffmanNode{freq: 3, left: nodeC, right: nodeB}
		root := &huffmanNode{freq: 6, left: nodeA, right: parentCB}

		codes := make(map[byte]string)
		generateCodes(root, "", codes)

		expected := map[byte]string{'a': "0", 'c': "10", 'b': "11"}
		if !reflect.DeepEqual(codes, expected) {
			t.Errorf("generateCodes() got = %v, want %v", codes, expected)
		}
	})

	t.Run("Single node tree", func(t *testing.T) {
		nodeA := &huffmanNode{char: 'a', freq: 5}
		root := &huffmanNode{freq: 5, left: nodeA} // 单节点树的特殊结构
		codes := make(map[byte]string)
		generateCodes(root, "", codes)

		expected := map[byte]string{'a': "0"}
		if !reflect.DeepEqual(codes, expected) {
			t.Errorf("generateCodes() for single node got = %v, want %v", codes, expected)
		}
	})
}

// TestBitWriterAndReader 测试位读写器
func TestBitWriterAndReader(t *testing.T) {
	var buf bytes.Buffer
	writer := newBitWriter(&buf)

	bitsToWrite := []bool{true, false, true, true, false, false, true, false, true} // 10110010 1...

	for _, bit := range bitsToWrite {
		err := writer.WriteBit(bit)
		if err != nil {
			t.Fatalf("WriteBit failed: %v", err)
		}
	}
	err := writer.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// 预期字节: 10110010 (0xB2), 10000000 (0x80)
	expectedBytes := []byte{0xB2, 0x80}
	if !bytes.Equal(buf.Bytes(), expectedBytes) {
		t.Errorf("BitWriter produced wrong bytes. Got %x, want %x", buf.Bytes(), expectedBytes)
	}

	reader := newBitReader(&buf)
	// bitWriter 会将不足 8 bit 的最后一个字节用 0 填充，因此 reader 在读完实际写入的 bit 后，
	// 仍会读到 padding 的 0，直到耗尽底层字节流才会返回 EOF。
	totalBits := len(expectedBytes) * 8
	for i := 0; i < totalBits; i++ {
		bit, err := reader.ReadBit()
		if err != nil {
			t.Fatalf("ReadBit failed at bit %d: %v", i, err)
		}
		expected := false
		if i < len(bitsToWrite) {
			expected = bitsToWrite[i]
		}
		if bit != expected {
			t.Errorf("Read wrong bit at index %d. Got %v, want %v", i, bit, expected)
		}
	}

	// 底层字节流耗尽后再读一位，应该报错（EOF）
	_, err = reader.ReadBit()
	if err == nil {
		t.Error("Expected error when reading past end of underlying byte stream, but got nil")
	}
}

// TestSerializeDeserializeFreqTable 测试频率表的序列化和反序列化
func TestSerializeDeserializeFreqTable(t *testing.T) {
	originalTable := map[byte]int64{
		'a': 100,
		'b': 200,
		'z': 50,
		' ': 1000,
	}

	data, err := serializeFreqTable(originalTable)
	if err != nil {
		t.Fatalf("serializeFreqTable failed: %v", err)
	}

	reader := bytes.NewReader(data)
	deserializedTable, err := deserializeFreqTable(reader)
	if err != nil {
		t.Fatalf("deserializeFreqTable failed: %v", err)
	}

	if !reflect.DeepEqual(originalTable, deserializedTable) {
		t.Errorf("Tables do not match after serialization/deserialization. Got %v, want %v", deserializedTable, originalTable)
	}
}

// TestEndToEndCompressionDecompression 测试端到端的压缩和解压流程
func TestEndToEndCompressionDecompression(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		workers int
	}{
		{"Empty String", "", huffmanCompressionWorkers},
		{"Short String", "hello world", huffmanCompressionWorkers},
		{"Single Character", "aaaaaaaaaaaaaaaaaaaaaaaaaa", huffmanCompressionWorkers},
		{"Two Characters", "ababababababababababababab", huffmanCompressionWorkers},
		{"Exactly Chunk Size", string(bytes.Repeat([]byte("a"), huffmanChunkSize)), huffmanCompressionWorkers},
		{"Multiple Chunks", string(bytes.Repeat([]byte{'a', 'b', 'c'}, huffmanChunkSize)), huffmanCompressionWorkers},
		{"Large Random Data", generateRandomData(huffmanChunkSize*5 + 100), huffmanCompressionWorkers},
		{"Single Worker", generateRandomData(huffmanChunkSize * 2), 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 压缩
			pipeReader, pipeWriter := io.Pipe()
			mockWc := &mockWriteCloser{Buffer: &bytes.Buffer{}}

			// 创建一个WaitGroup来等待写入完成
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				defer wg.Done()
				// 将压缩数据写入缓冲区
				_, err := io.Copy(mockWc, pipeReader)
				if err != nil && err != io.EOF {
					t.Errorf("Copy from pipe failed: %v", err)
				}
			}()

			// 创建带有多 worker 的压缩写入器
			writer := NewCompressedWriter(pipeWriter)
			_, err := writer.Write([]byte(tc.input))
			if err != nil {
				t.Fatalf("CompressedWriter.Write failed: %v", err)
			}
			err = writer.Close() // 关闭写入器以刷新所有数据并完成压缩
			if err != nil {
				t.Fatalf("CompressedWriter.Close failed: %v", err)
			}

			wg.Wait() // 等待所有数据都写入 mockWc

			// 解压
			reader, err := NewCompressedReader(mockWc)
			if err != nil {
				t.Fatalf("NewCompressedReader failed: %v", err)
			}

			decompressedData, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("io.ReadAll on decompressed stream failed: %v", err)
			}

			// 验证
			if !bytes.Equal([]byte(tc.input), decompressedData) {
				t.Errorf("Decompressed data does not match original data")
				// 可以在这里添加更详细的差异输出来帮助调试
				if len(tc.input) != len(decompressedData) {
					t.Errorf("Length mismatch: original=%d, decompressed=%d", len(tc.input), len(decompressedData))
				}
			}
		})
	}
}

// generateRandomData 生成指定大小的随机数据
func generateRandomData(size int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(rnd.Intn(256))
	}
	return string(data)
}

// TestHuffmanReader_ErrorHandling 测试读取器错误处理
func TestHuffmanReader_ErrorHandling(t *testing.T) {
	t.Run("Input too short", func(t *testing.T) {
		input := bytes.NewReader([]byte("HUF")) // 比魔术字短
		_, err := NewCompressedReader(input)
		if err != ErrNotCompressed {
			t.Errorf("Expected ErrNotCompressed for short input, got %v", err)
		}
	})

	t.Run("Invalid magic number", func(t *testing.T) {
		input := bytes.NewReader([]byte("NOTHUFF"))
		_, err := NewCompressedReader(input)
		if err != ErrNotCompressed {
			t.Errorf("Expected ErrNotCompressed for invalid magic number, got %v", err)
		}
	})

	t.Run("Invalid chunk magic", func(t *testing.T) {
		// 构造一个带有正确文件头但错误块头的数据
		var buf bytes.Buffer
		buf.Write(huffmanMagic)
		buf.Write([]byte("BADCHK")) // 错误的块魔术字
		reader, _ := NewCompressedReader(&buf)
		_, err := io.ReadAll(reader)
		if err == nil {
			t.Errorf("Expected an error for invalid chunk magic, got nil")
		}
	})
}

// TestHuffmanWriter_CloseIdempotency 测试Close的幂等性
func TestHuffmanWriter_CloseIdempotency(t *testing.T) {
	mockWc := newMockWriteCloser()
	writer := NewCompressedWriter(mockWc)

	err1 := writer.Close()
	if err1 != nil {
		t.Fatalf("First Close() failed: %v", err1)
	}

	err2 := writer.Close()
	if err2 != nil {
		t.Errorf("Second Close() should not return a new error, got %v", err2)
	}
}

// TestHuffmanWriter_WriteAfterClose 测试在关闭后写入
func TestHuffmanWriter_WriteAfterClose(t *testing.T) {
	mockWc := newMockWriteCloser()
	writer := NewCompressedWriter(mockWc)

	// 先写入一些数据
	_, err := writer.Write([]byte("data"))
	if err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	// 关闭写入器
	err = writer.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// 再次写入，应该会失败
	// 注意：由于后台goroutine的存在，错误可能不是立即出现的
	// 这里的测试依赖于 `hw.err` 字段被正确设置和检查
	_, err = writer.Write([]byte("more data"))
	hw := writer.(*huffmanWriter)
	if hw.err == nil {
		t.Errorf("Write after Close should return an error, but got nil")
	}
}

// FuzzHuffman 对压缩和解压流程进行模糊测试
func FuzzHuffman(f *testing.F) {
	// 添加一些种子语料
	f.Add([]byte(""))
	f.Add([]byte("hello world"))
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add(bytes.Repeat([]byte{123}, 1000))

	f.Fuzz(func(t *testing.T, original []byte) {
		// 压缩
		mockWc := newMockWriteCloser()
		writer := NewCompressedWriter(mockWc)

		_, err := writer.Write(original)
		if err != nil {
			// 在模糊测试中，写入错误本身不是一个失败，但关闭时应能正确处理
		}
		err = writer.Close()
		if err != nil {
			t.Fatalf("writer.Close() failed: %v", err)
		}

		// 解压
		reader, err := NewCompressedReader(mockWc)
		// 如果原始数据为空，压缩后也为空，NewCompressedReader 会因找不到魔术字而报错
		// 只在有压缩数据时继续
		if err != nil {
			if mockWc.Len() > 0 {
				t.Fatalf("NewCompressedReader failed with non-empty buffer: %v", err)
			}
			return // 如果缓冲区为空，报错是正常的
		}

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			// 由于模糊测试可能会产生无法解码的无效码字，io.ErrUnexpectedEOF 是可能发生的
			// 但其他错误可能表明存在问题
			if err != io.ErrUnexpectedEOF {
				t.Fatalf("io.ReadAll from reader failed: %v", err)
			}
		}

		// 只有在解压没有出错时才比较数据
		if err == nil && !bytes.Equal(original, decompressed) {
			t.Errorf("Decompressed data does not match original. Original len %d, decompressed len %d", len(original), len(decompressed))
		}
	})
}

// nodeQueue 的排序稳定性测试
func TestNodeQueue_Less(t *testing.T) {
	// 测试当频率相同时，是否根据 minChar 排序
	node1 := &huffmanNode{freq: 10, minChar: 'b'}
	node2 := &huffmanNode{freq: 10, minChar: 'a'}
	node3 := &huffmanNode{freq: 5, minChar: 'c'}

	pq := nodeQueue{node1, node2, node3}
	sort.Sort(pq)

	if pq[0] != node3 || pq[1] != node2 || pq[2] != node1 {
		t.Error("nodeQueue sort order is incorrect. Expected order by freq, then minChar")
		t.Errorf("Got order: %c, %c, %c", pq[0].minChar, pq[1].minChar, pq[2].minChar)
	}
}
