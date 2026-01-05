// core/huffman.go
package core

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
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
var huffmanMagic = []byte("HUFF") // 标识流的开始
var chunkMagic = []byte("HCHK")   // 标识每个块

const (
	// 定义压缩块的大小
	huffmanChunkSize = 256 * 1024 // 256 KB
	maxHuffmanChunkLen = 4 * 1024 * 1024
	maxHuffmanHeaderLen = 4096
)

var (
	// 定义用于压缩的 worker 数量
	huffmanCompressionWorkers = runtime.NumCPU()
)

// huffmanJob 包含一个待压缩的数据块
type huffmanJob struct {
	id   int
	data []byte
}

// huffmanResult 包含一个已压缩的数据块
type huffmanResult struct {
	id   int
	data []byte
	err  error
}

type huffmanWriter struct {
	w io.WriteCloser

	buffer *bytes.Buffer

	jobs     chan huffmanJob
	results  chan huffmanResult
	wg       sync.WaitGroup // For compressWorker goroutines
	writerWg sync.WaitGroup // For resultWriter
	errOnce  sync.Once
	err      error

	nextID int
	closed atomic.Bool // <<-- 重入保护
}

func NewCompressedWriter(w io.WriteCloser) io.WriteCloser {
	hw := &huffmanWriter{
		w:       w,
		buffer:  bytes.NewBuffer(make([]byte, 0, huffmanChunkSize)),
		jobs:    make(chan huffmanJob, huffmanCompressionWorkers),
		results: make(chan huffmanResult, huffmanCompressionWorkers),
	}

	// 启动结果写入器
	hw.writerWg.Add(1)
	go hw.resultWriter()

	// 启动压缩 worker
	for i := 0; i < huffmanCompressionWorkers; i++ {
		hw.wg.Add(1)
		go hw.compressWorker()
	}

	if _, err := hw.w.Write(huffmanMagic); err != nil {
		hw.setError(err)
	}
	return hw
}

// compressWorker 是一个后台 goroutine，用于压缩数据块
func (hw *huffmanWriter) compressWorker() {
	defer hw.wg.Done()
	for job := range hw.jobs {
		compressedData, err := compressChunk(job.data)
		if err != nil {
			hw.setError(err)
			hw.results <- huffmanResult{id: job.id, err: err}
			return
		}
		hw.results <- huffmanResult{id: job.id, data: compressedData}
	}
}

// resultWriter 是一个后台 goroutine，用于按顺序将压缩结果写入底层 writer
func (hw *huffmanWriter) resultWriter() {
	defer hw.writerWg.Done()

	pending := make(map[int][]byte)
	nextID := 0
	for result := range hw.results {
		if result.err != nil {
			hw.setError(result.err)
			continue // 继续处理以清空通道，防止 worker 阻塞
		}

		pending[result.id] = result.data

		// 尝试按顺序写入
		for {
			data, ok := pending[nextID]
			if !ok {
				break // 下一个块还没到
			}
			// 写入块魔术字
			if _, err := hw.w.Write(chunkMagic); err != nil {
				hw.setError(err)
				return
			}
			// 写入压缩数据块的长度
			if err := binary.Write(hw.w, binary.BigEndian, uint32(len(data))); err != nil {
				hw.setError(err)
				return
			}
			// 写入压缩数据
			if _, err := hw.w.Write(data); err != nil {
				hw.setError(err)
				return
			}
			delete(pending, nextID)
			nextID++
		}
	}
}

// compressChunk 压缩单个数据块
func compressChunk(originalData []byte) ([]byte, error) {
	originalLen := uint64(len(originalData))
	if originalLen == 0 {
		return []byte{}, nil
	}

	freqTable := buildFrequencyTable(originalData)
	tree := buildHuffmanTree(freqTable)
	codes := make(map[byte]string)
	generateCodes(tree, "", codes)

	headerBytes, err := serializeFreqTable(freqTable)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal huffman frequency table: %w", err)
	}

	var buf bytes.Buffer
	// 写入原始数据长度
	if err := binary.Write(&buf, binary.BigEndian, originalLen); err != nil {
		return nil, err
	}
	// 写入频率表长度
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(headerBytes))); err != nil {
		return nil, err
	}
	// 写入频率表
	if _, err := buf.Write(headerBytes); err != nil {
		return nil, err
	}

	// 写入压缩数据
	bw := newBitWriter(&buf)
	for _, b := range originalData {
		code := codes[b]
		for _, bitChar := range code {
			if err := bw.WriteBit(bitChar == '1'); err != nil {
				return nil, err
			}
		}
	}
	if err := bw.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (hw *huffmanWriter) setError(err error) {
	if err != nil {
		hw.errOnce.Do(func() {
			hw.err = err
		})
	}
}

func (hw *huffmanWriter) Write(p []byte) (int, error) {
	if hw.closed.Load() {
		hw.setError(ErrWriterClosed)
		return 0, ErrWriterClosed
	}
	if hw.err != nil {
		return 0, hw.err
	}

	written := 0
	for len(p) > 0 {
		space := huffmanChunkSize - hw.buffer.Len()
		if space == 0 {
			if err := hw.flush(false); err != nil {
				return written, err
			}
			space = huffmanChunkSize
		}

		toWrite := len(p)
		if toWrite > space {
			toWrite = space
		}

		n, _ := hw.buffer.Write(p[:toWrite])
		written += n
		p = p[n:]
	}

	return written, nil
}

// flush 将当前缓冲区的数据发送去压缩
func (hw *huffmanWriter) flush(final bool) error {
	if hw.err != nil {
		return hw.err
	}
	if hw.buffer.Len() == 0 && !final {
		return nil
	}

	dataToCompress := make([]byte, hw.buffer.Len())
	copy(dataToCompress, hw.buffer.Bytes())
	hw.buffer.Reset()

	hw.jobs <- huffmanJob{id: hw.nextID, data: dataToCompress}
	hw.nextID++

	return nil
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
	// 如果已经关闭，则直接返回错误
	if !hw.closed.CompareAndSwap(false, true) {
		return hw.err
	}

	// 总是尝试关闭底层 writer
	defer hw.w.Close()

	if err := hw.flush(true); err != nil {
		hw.setError(err)
	}

	// 关闭 jobs channel，通知 worker 没有更多任务
	close(hw.jobs)

	// 等待所有 compress worker 完成
	hw.wg.Wait()

	// 所有 worker 完成后，关闭 results channel
	close(hw.results)

	// 等待 resultWriter 完成
	hw.writerWg.Wait()

	return hw.err
}

// --- huffmanReader ---
var ErrNotCompressed = errors.New("not a huffman compressed file")
var ErrWriterClosed = errors.New("huffman writer is closed")

type huffmanReader struct {
	r      io.Reader
	buffer *bytes.Buffer // 存储解压后的数据
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

func NewCompressedReader(r io.Reader) (io.ReadCloser, error) {
	return NewParallelCompressedReader(r)
}

// --- 并行解压缩 ---

var huffmanDecompressionWorkers = runtime.NumCPU()

type decompressJob struct {
	id   int
	data []byte
}

type decompressResult struct {
	id   int
	data []byte
	err  error
}

func NewParallelCompressedReader(r io.Reader) (io.ReadCloser, error) {
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

	pr, pw := io.Pipe()

	go func() {
		var pipelineErr error
		defer func() {
			if pipelineErr != nil {
				pw.CloseWithError(pipelineErr)
			} else {
				pw.Close()
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		jobs := make(chan decompressJob, huffmanDecompressionWorkers)
		results := make(chan decompressResult, huffmanDecompressionWorkers)
		var wg sync.WaitGroup

		// 启动 workers
		wg.Add(huffmanDecompressionWorkers)
		for i := 0; i < huffmanDecompressionWorkers; i++ {
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case j, ok := <-jobs:
						if !ok {
							return
						}
						decompressed, err := decompressChunk(j.data)
						select {
						case results <- decompressResult{id: j.id, data: decompressed, err: err}:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		// 启动 producer
		producerWg := sync.WaitGroup{}
		producerWg.Add(1)
		go func() {
			defer producerWg.Done()
			defer close(jobs)
			nextID := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				chunkHeader := make([]byte, len(chunkMagic)+4) // magic + len
				if _, err := io.ReadFull(r, chunkHeader); err != nil {
					if err == io.EOF {
						return // 正常结束
					}
					pipelineErr = fmt.Errorf("failed to read huffman chunk header: %w", err)
					cancel()
					return
				}

				if !bytes.Equal(chunkHeader[:len(chunkMagic)], chunkMagic) {
					pipelineErr = fmt.Errorf("invalid huffman chunk magic")
					cancel()
					return
				}

				chunkLen := binary.BigEndian.Uint32(chunkHeader[len(chunkMagic):])
				if chunkLen == 0 {
					return // 0-length chunk as end marker
				}
				if chunkLen > maxHuffmanChunkLen {
					pipelineErr = fmt.Errorf("huffman chunk too large: %d", chunkLen)
					cancel()
					return
				}
				chunkData := make([]byte, chunkLen)
				if _, err := io.ReadFull(r, chunkData); err != nil {
					pipelineErr = fmt.Errorf("failed to read huffman chunk data: %w", err)
					cancel()
					return
				}

				select {
				case jobs <- decompressJob{id: nextID, data: chunkData}:
					nextID++
				case <-ctx.Done():
					return
				}
			}
		}()

		go func() {
			producerWg.Wait()
			wg.Wait()
			close(results)
		}()

		// Aggregator
		resultsBuffer := make(map[int][]byte)
		nextWriteID := 0
		for {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-results:
				if !ok {
					return // 正常结束
				}
				if res.err != nil {
					pipelineErr = res.err
					cancel()
					continue
				}

				resultsBuffer[res.id] = res.data
				for {
					data, exists := resultsBuffer[nextWriteID]
					if !exists {
						break
					}

					if _, err := pw.Write(data); err != nil {
						cancel()
						return
					}
					delete(resultsBuffer, nextWriteID)
					nextWriteID++
				}
			}
		}
	}()

	return pr, nil
}

// decompressChunk 将单个压缩数据块解压为原始数据
func decompressChunk(chunkData []byte) ([]byte, error) {
	chunkReader := bytes.NewReader(chunkData)

	var originalLen uint64
	if err := binary.Read(chunkReader, binary.BigEndian, &originalLen); err != nil {
		return nil, err
	}
	if originalLen == 0 {
		return []byte{}, nil
	}
	if originalLen > uint64(huffmanChunkSize) {
		return nil, fmt.Errorf("invalid huffman original length: %d", originalLen)
	}

	var headerLen uint32
	if err := binary.Read(chunkReader, binary.BigEndian, &headerLen); err != nil {
		return nil, err
	}
	if headerLen == 0 || headerLen > maxHuffmanHeaderLen {
		return nil, fmt.Errorf("invalid huffman header length: %d", headerLen)
	}

	freqTable, err := deserializeFreqTable(io.LimitReader(chunkReader, int64(headerLen)))
	if err != nil {
		return nil, err
	}

	decodeTree := buildHuffmanTree(freqTable)
	if decodeTree == nil {
		return nil, fmt.Errorf("failed to build huffman tree from chunk")
	}

	bitReader := newBitReader(chunkReader)
	out := make([]byte, 0, originalLen)

	for i := uint64(0); i < originalLen; i++ {
		currentNode := decodeTree
		for currentNode.left != nil || currentNode.right != nil {
			bit, err := bitReader.ReadBit()
			if err != nil {
				return nil, io.ErrUnexpectedEOF
			}
			if bit {
				currentNode = currentNode.right
			} else {
				currentNode = currentNode.left
			}
			if currentNode == nil {
				return nil, fmt.Errorf("invalid huffman code in chunk")
			}
		}
		out = append(out, currentNode.char)
	}

	return out, nil
}

func (hr *huffmanReader) Read(p []byte) (n int, err error) {
	// 首先从缓冲区读取
	if hr.buffer.Len() > 0 {
		return hr.buffer.Read(p)
	}

	// 缓冲区为空，解压下一个块
	err = hr.decompressNextChunk()
	if err != nil {
		return 0, err // 包括 io.EOF
	}

	// 再次尝试从缓冲区读取
	return hr.buffer.Read(p)
}

func (hr *huffmanReader) decompressNextChunk() error {
	// 读取块魔术字
	magic := make([]byte, len(chunkMagic))
	if _, err := io.ReadFull(hr.r, magic); err != nil {
		return err // 可能是 EOF，表示流结束
	}
	if !bytes.Equal(magic, chunkMagic) {
		return fmt.Errorf("invalid huffman chunk magic")
	}

	// 读取块长度
	var chunkLen uint32
	if err := binary.Read(hr.r, binary.BigEndian, &chunkLen); err != nil {
		return err
	}
	if chunkLen > maxHuffmanChunkLen {
		return fmt.Errorf("huffman chunk too large: %d", chunkLen)
	}

	// 读取整个压缩块
	chunkData := make([]byte, chunkLen)
	if _, err := io.ReadFull(hr.r, chunkData); err != nil {
		return err
	}

	chunkReader := bytes.NewReader(chunkData)

	// 解压块
	var originalLen uint64
	if err := binary.Read(chunkReader, binary.BigEndian, &originalLen); err != nil {
		return err
	}
	if originalLen == 0 {
		return nil // 空块
	}
	if originalLen > uint64(huffmanChunkSize) {
		return fmt.Errorf("invalid huffman original length: %d", originalLen)
	}

	var headerLen uint32
	if err := binary.Read(chunkReader, binary.BigEndian, &headerLen); err != nil {
		return err
	}
	if headerLen == 0 || headerLen > maxHuffmanHeaderLen {
		return fmt.Errorf("invalid huffman header length: %d", headerLen)
	}

	freqTable, err := deserializeFreqTable(io.LimitReader(chunkReader, int64(headerLen)))
	if err != nil {
		return err
	}

	decodeTree := buildHuffmanTree(freqTable)
	if decodeTree == nil {
		return fmt.Errorf("failed to build huffman tree from chunk")
	}

	bitReader := newBitReader(chunkReader)

	// 重置缓冲区并解压数据到其中
	hr.buffer.Reset()
	hr.buffer.Grow(int(originalLen))

	for i := uint64(0); i < originalLen; i++ {
		currentNode := decodeTree
		for currentNode.left != nil || currentNode.right != nil {
			bit, err := bitReader.ReadBit()
			if err != nil {
				if err == io.EOF {
					return io.ErrUnexpectedEOF
				}
				return err
			}
			if bit {
				currentNode = currentNode.right
			} else {
				currentNode = currentNode.left
			}
			if currentNode == nil {
				return fmt.Errorf("invalid huffman code in chunk")
			}
		}
		hr.buffer.WriteByte(currentNode.char)
	}

	return nil
}
