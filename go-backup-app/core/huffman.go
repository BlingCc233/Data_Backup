// core/huffman.go
package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
)

// huffmanNode 表示Huffman树中的一个节点
type huffmanNode struct {
	char    byte
	freq    int64
	left    *huffmanNode
	right   *huffmanNode
	minChar byte // 用于在频率相同时作为排序的次要依据
}

// nodeQueue 是用于构建树的优先队列
type nodeQueue []*huffmanNode

func (nq nodeQueue) Len() int      { return len(nq) }
func (nq nodeQueue) Swap(i, j int) { nq[i], nq[j] = nq[j], nq[i] }
func (nq nodeQueue) Less(i, j int) bool {
	if nq[i].freq < nq[j].freq {
		return true
	}
	if nq[i].freq > nq[j].freq {
		return false
	}
	return nq[i].minChar < nq[j].minChar
}

// buildFrequencyTable 计算输入数据的字节频率
func buildFrequencyTable(data []byte) map[byte]int64 {
	freqTable := make(map[byte]int64)
	for _, b := range data {
		freqTable[b]++
	}
	return freqTable
}

// buildHuffmanTree 根据频率表构建Huffman树
func buildHuffmanTree(freqTable map[byte]int64) *huffmanNode {
	if len(freqTable) == 0 {
		return nil
	}
	var pq nodeQueue
	for char, freq := range freqTable {
		pq = append(pq, &huffmanNode{char: char, freq: freq, minChar: char})
	}
	if len(pq) == 1 {
		root := &huffmanNode{freq: pq[0].freq, left: pq[0], minChar: pq[0].minChar}
		return root
	}
	for len(pq) > 1 {
		sort.Sort(pq)
		left := pq[0]
		right := pq[1]
		pq = pq[2:]
		parent := &huffmanNode{
			freq:    left.freq + right.freq,
			left:    left,
			right:   right,
			minChar: left.minChar,
		}
		pq = append(pq, parent)
	}
	return pq[0]
}

// generateCodes 递归遍历Huffman树以生成编码表
func generateCodes(node *huffmanNode, prefix string, codes map[byte]string) {
	if node == nil {
		return
	}
	if node.left == nil && node.right == nil { // Leaf node
		if prefix == "" {
			codes[node.char] = "0"
		} else {
			codes[node.char] = prefix
		}
		return
	}
	generateCodes(node.left, prefix+"0", codes)
	generateCodes(node.right, prefix+"1", codes)
}

// --- bitWriter ---
type bitWriter struct {
	w      io.Writer
	buffer byte
	count  uint8
}

func newBitWriter(w io.Writer) *bitWriter { return &bitWriter{w: w} }
func (bw *bitWriter) WriteBit(bit bool) error {
	if bit {
		bw.buffer |= (1 << (7 - bw.count))
	}
	bw.count++
	if bw.count == 8 {
		if _, err := bw.w.Write([]byte{bw.buffer}); err != nil {
			return err
		}
		bw.buffer = 0
		bw.count = 0
	}
	return nil
}
func (bw *bitWriter) Flush() error {
	if bw.count > 0 {
		if _, err := bw.w.Write([]byte{bw.buffer}); err != nil {
			return err
		}
	}
	return nil
}

// --- huffmanWriter ---
var huffmanMagic = []byte("HUFF")

type huffmanWriter struct {
	w      io.WriteCloser
	buffer *bytes.Buffer
}

func NewCompressedWriter(w io.WriteCloser) *huffmanWriter {
	return &huffmanWriter{w: w, buffer: new(bytes.Buffer)}
}
func (hw *huffmanWriter) Write(p []byte) (int, error) {
	return hw.buffer.Write(p)
}

func serializeFreqTable(freqTable map[byte]int64) ([]byte, error) {
	var buf bytes.Buffer

	// Write number of entries
	if err := binary.Write(&buf, binary.BigEndian, uint16(len(freqTable))); err != nil {
		return nil, err
	}

	// Sort keys to ensure deterministic output
	keys := make([]byte, 0, len(freqTable))
	for k := range freqTable {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, char := range keys {
		// Write character
		if err := buf.WriteByte(char); err != nil {
			return nil, err
		}
		// Write frequency
		if err := binary.Write(&buf, binary.BigEndian, uint64(freqTable[char])); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (hw *huffmanWriter) Close() error {
	defer hw.w.Close()

	originalData := hw.buffer.Bytes()
	originalLen := uint64(len(originalData))

	if originalLen == 0 {
		if _, err := hw.w.Write(huffmanMagic); err != nil {
			return err
		}
		if err := binary.Write(hw.w, binary.BigEndian, uint32(0)); err != nil {
			return err
		} // Header length
		if err := binary.Write(hw.w, binary.BigEndian, uint64(0)); err != nil {
			return err
		} // Original data length
		return nil
	}

	freqTable := buildFrequencyTable(originalData)
	tree := buildHuffmanTree(freqTable)
	codes := make(map[byte]string)
	generateCodes(tree, "", codes)

	headerBytes, err := serializeFreqTable(freqTable)
	if err != nil {
		return fmt.Errorf("failed to marshal huffman frequency table: %w", err)
	}
	headerLen := uint32(len(headerBytes))

	if _, err := hw.w.Write(huffmanMagic); err != nil {
		return fmt.Errorf("failed to write huffman magic: %w", err)
	}
	if err := binary.Write(hw.w, binary.BigEndian, headerLen); err != nil {
		return fmt.Errorf("failed to write huffman header length: %w", err)
	}
	if _, err := hw.w.Write(headerBytes); err != nil {
		return fmt.Errorf("failed to write huffman header: %w", err)
	}
	if err := binary.Write(hw.w, binary.BigEndian, originalLen); err != nil {
		return fmt.Errorf("failed to write original data length: %w", err)
	}
	bw := newBitWriter(hw.w)
	for _, b := range originalData {
		code := codes[b]
		for _, bitChar := range code {
			if err := bw.WriteBit(bitChar == '1'); err != nil {
				return fmt.Errorf("failed to write compressed bit: %w", err)
			}
		}
	}

	return bw.Flush()
}

// --- huffmanReader ---
var ErrNotCompressed = errors.New("not a huffman compressed file")

type huffmanReader struct {
	r          io.Reader
	bitReader  *bitReader
	decodeTree *huffmanNode
	decoded    uint64
	totalSize  uint64
}

type bitReader struct {
	r      io.Reader
	buffer byte
	count  uint8
}

func newBitReader(r io.Reader) *bitReader { return &bitReader{r: r} }
func (br *bitReader) ReadBit() (bool, error) {
	if br.count == 0 {
		buf := make([]byte, 1)
		n, err := br.r.Read(buf)
		if err != nil {
			return false, err
		}
		if n == 0 {
			return false, io.EOF
		}
		br.buffer = buf[0]
		br.count = 8
	}
	br.count--
	bit := (br.buffer & (1 << br.count)) != 0
	return bit, nil
}

// deserializeFreqTable reads the binary header and reconstructs the frequency table.
func deserializeFreqTable(r io.Reader) (map[byte]int64, error) {
	var numEntries uint16
	if err := binary.Read(r, binary.BigEndian, &numEntries); err != nil {
		return nil, fmt.Errorf("failed to read freq table entry count: %w", err)
	}

	freqTable := make(map[byte]int64)
	for i := 0; i < int(numEntries); i++ {
		charBuf := make([]byte, 1)
		if _, err := io.ReadFull(r, charBuf); err != nil {
			return nil, fmt.Errorf("failed to read char in freq table: %w", err)
		}
		char := charBuf[0]

		var freq uint64
		if err := binary.Read(r, binary.BigEndian, &freq); err != nil {
			return nil, fmt.Errorf("failed to read freq in freq table: %w", err)
		}
		freqTable[char] = int64(freq)
	}
	return freqTable, nil
}

func NewCompressedReader(r io.Reader) (io.Reader, error) {
	magic := make([]byte, len(huffmanMagic))
	if _, err := io.ReadFull(r, magic); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, ErrNotCompressed
		}
		return nil, err
	}
	if !bytes.Equal(magic, huffmanMagic) {
		return nil, ErrNotCompressed
	}

	var headerLen uint32
	if err := binary.Read(r, binary.BigEndian, &headerLen); err != nil {
		return nil, fmt.Errorf("failed to read huffman header length: %w", err)
	}

	if headerLen == 0 {
		var originalLen uint64
		if err := binary.Read(r, binary.BigEndian, &originalLen); err != nil {
			return nil, fmt.Errorf("failed to read original length for empty file: %w", err)
		}
		if originalLen == 0 {
			return bytes.NewReader([]byte{}), nil
		}
		return nil, fmt.Errorf("invalid empty huffman header")
	}

	// Read the entire header into a buffer to pass to deserializeFreqTable
	headerBytes := make([]byte, headerLen)
	if _, err := io.ReadFull(r, headerBytes); err != nil {
		return nil, fmt.Errorf("failed to read huffman header: %w", err)
	}
	headerReader := bytes.NewReader(headerBytes)

	freqTable, err := deserializeFreqTable(headerReader)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal huffman frequency table: %w", err)
	}

	var originalLen uint64
	if err := binary.Read(r, binary.BigEndian, &originalLen); err != nil {
		return nil, fmt.Errorf("failed to read original data length: %w", err)
	}

	decodeTree := buildHuffmanTree(freqTable)

	return &huffmanReader{
		r:          r,
		bitReader:  newBitReader(r),
		decodeTree: decodeTree,
		totalSize:  originalLen,
	}, nil
}

func (hr *huffmanReader) Read(p []byte) (n int, err error) {
	if hr.decoded >= hr.totalSize {
		return 0, io.EOF
	}
	if hr.decodeTree == nil {
		return 0, io.EOF
	}
	for n < len(p) {
		if hr.decoded >= hr.totalSize {
			break
		}
		currentNode := hr.decodeTree
		for currentNode.left != nil || currentNode.right != nil {
			bit, err := hr.bitReader.ReadBit()
			if err != nil {
				if err == io.EOF && hr.decoded < hr.totalSize {
					return n, io.ErrUnexpectedEOF
				}
				return n, err
			}
			if bit {
				currentNode = currentNode.right
			} else {
				currentNode = currentNode.left
			}
			if currentNode == nil {
				return n, fmt.Errorf("invalid huffman code encountered")
			}
		}
		p[n] = currentNode.char
		n++
		hr.decoded++
	}
	if hr.decoded >= hr.totalSize {
		return n, io.EOF
	}
	return n, nil
}
