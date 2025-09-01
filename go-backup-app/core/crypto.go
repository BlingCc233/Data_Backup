// core/crypto.go
package core

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256" // 我们用标准库的SHA256来做KDF，自己实现SHA256过于庞大
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/crypto/pbkdf2" // 同样，自己实现PBKDF2非常复杂，这里使用x/crypto
	"io"
)

// 🚨 安全警告: 以下AES和ChaCha20的实现仅为教学目的，请勿用于生产环境！
// 真实项目中请使用 Go 的 `crypto/aes` 和 `golang.org/x/crypto/chacha20poly1305`。

var (
	magicHeader         = []byte("QBAKENCR")
	ErrInvalidMagic     = errors.New("invalid magic header: not an encrypted qbak file")
	ErrPasswordRequired = errors.New("password is required for this encrypted file")
)

const (
	version        = 0x01
	AlgoAES256_CTR = 0x01
	AlgoChaCha20   = 0x02
)

// --- 密钥派生 ---
// 我们使用标准、安全的PBKDF2，而不是自己实现。这是密码学中的最佳实践。
func deriveKey(password string, salt []byte) []byte {
	// 使用 32 字节 (256位) 的密钥长度
	return pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
}

// --- 加密写入器 ---
type encryptedWriter struct {
	underlyingWriter io.Writer
	streamCipher     io.Writer // 这将是我们的加密流
	buf              *bytes.Buffer
}

func NewEncryptedWriter(w io.Writer, password string, algorithm uint8) (io.WriteCloser, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty for encryption")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := deriveKey(password, salt)

	var nonce []byte
	var stream io.Writer

	switch algorithm {
	case AlgoAES256_CTR:
		nonce = make([]byte, 16) // AES IV is 16 bytes
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}
		// 这是一个简化的 AES-CTR 流写入器
		aesStream, err := NewSimpleAESCTRStream(nil, key, nonce)
		if err != nil {
			return nil, err
		}
		stream = aesStream

	case AlgoChaCha20:
		nonce = make([]byte, 12) // ChaCha20 nonce is 12 bytes
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}
		// 简化的 ChaCha20 流写入器
		chachaStream, err := NewSimpleChaCha20Stream(nil, key, nonce)
		if err != nil {
			return nil, err
		}
		stream = chachaStream

	default:
		return nil, fmt.Errorf("unsupported algorithm: %d", algorithm)
	}

	// 写入加密文件头
	header := new(bytes.Buffer)
	header.Write(magicHeader)
	header.WriteByte(version)
	header.WriteByte(byte(algorithm))
	header.WriteByte(byte(len(salt)))
	header.Write(salt)
	header.WriteByte(byte(len(nonce)))
	header.Write(nonce)

	if _, err := w.Write(header.Bytes()); err != nil {
		return nil, err
	}

	// 将底层写入器和加密流关联起来
	// 我们需要一个中间缓冲区来处理数据块
	buf := new(bytes.Buffer)
	if s, ok := stream.(*SimpleAESCTRStream); ok {
		s.w = w
	}
	if s, ok := stream.(*SimpleChaCha20Stream); ok {
		s.w = w
	}

	return &encryptedWriter{
		underlyingWriter: w,
		streamCipher:     stream,
		buf:              buf,
	}, nil
}

func (ew *encryptedWriter) Write(p []byte) (n int, err error) {
	// 数据写入到我们的加密流中，它内部会处理加密并写入到底层writer
	return ew.streamCipher.Write(p)
}

func (ew *encryptedWriter) Close() error {
	// 确保所有缓冲数据都被写入
	if c, ok := ew.streamCipher.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// --- 解密读取器 ---

type decryptedReader struct {
	streamCipher io.Reader
}

func NewDecryptedReader(r io.Reader, password string) (io.Reader, error) {
	header := make([]byte, len(magicHeader))
	if _, err := io.ReadFull(r, header); err != nil {
		if err == io.EOF {
			return nil, ErrInvalidMagic // 文件太短
		}
		return nil, err
	}
	if !bytes.Equal(header, magicHeader) {
		return nil, ErrInvalidMagic
	}

	// 如果是加密文件但没有密码，返回特定错误
	if password == "" {
		return nil, ErrPasswordRequired
	}

	// 读取版本和算法
	meta := make([]byte, 2)
	if _, err := io.ReadFull(r, meta); err != nil {
		return nil, err
	}
	// versionByte := meta[0]
	algoByte := meta[1]

	// 读取 Salt
	saltLen := make([]byte, 1)
	if _, err := io.ReadFull(r, saltLen); err != nil {
		return nil, err
	}
	salt := make([]byte, saltLen[0])
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, err
	}

	// 读取 Nonce/IV
	nonceLen := make([]byte, 1)
	if _, err := io.ReadFull(r, nonceLen); err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceLen[0])
	if _, err := io.ReadFull(r, nonce); err != nil {
		return nil, err
	}

	key := deriveKey(password, salt)

	var stream io.Reader
	switch algoByte {
	case AlgoAES256_CTR:
		aesStream, err := NewSimpleAESCTRStream(r, key, nonce)
		if err != nil {
			return nil, err
		}
		stream = aesStream
	case AlgoChaCha20:
		chachaStream, err := NewSimpleChaCha20Stream(r, key, nonce)
		if err != nil {
			return nil, err
		}
		stream = chachaStream
	default:
		return nil, fmt.Errorf("unsupported algorithm: %d", algoByte)
	}

	return &decryptedReader{streamCipher: stream}, nil
}

func (dr *decryptedReader) Read(p []byte) (n int, err error) {
	return dr.streamCipher.Read(p)
}

// --- 简化版 AES-CTR 实现 (仅教学) ---
// 这是一个高度简化的、不完整的 AES 实现，仅用于演示CTR模式。
type SimpleAESCTRStream struct {
	r       io.Reader
	w       io.Writer
	key     []byte
	iv      []byte
	counter uint64
}

func NewSimpleAESCTRStream(r io.Reader, key, iv []byte) (*SimpleAESCTRStream, error) {
	if len(key) != 32 || len(iv) != 16 {
		return nil, errors.New("invalid key/iv length for AES-256")
	}
	return &SimpleAESCTRStream{r: r, key: key, iv: iv}, nil
}

// 伪加密块函数，实际应调用完整AES
func (s *SimpleAESCTRStream) pseudoEncryptBlock(block []byte) []byte {
	// 这只是一个占位符！实际的AES非常复杂。
	// 我们用一个简单的XOR来模拟加密效果。
	result := make([]byte, len(block))
	for i := range block {
		result[i] = block[i] ^ s.key[i%len(s.key)]
	}
	return result
}

func (s *SimpleAESCTRStream) getKeystreamBlock() []byte {
	counterBlock := make([]byte, 16)
	copy(counterBlock, s.iv)
	binary.BigEndian.PutUint64(counterBlock[8:], s.counter)
	s.counter++
	return s.pseudoEncryptBlock(counterBlock)
}

func (s *SimpleAESCTRStream) Read(p []byte) (n int, err error) {
	// 简化读取逻辑
	underlyingData := make([]byte, len(p))
	n, err = s.r.Read(underlyingData)
	if n > 0 {
		keystream := s.getKeystreamBlock()
		for i := 0; i < n; i++ {
			p[i] = underlyingData[i] ^ keystream[i%16]
		}
	}
	return n, err
}

func (s *SimpleAESCTRStream) Write(p []byte) (n int, err error) {
	// 简化写入逻辑
	encryptedData := make([]byte, len(p))
	keystream := s.getKeystreamBlock()
	for i := 0; i < len(p); i++ {
		encryptedData[i] = p[i] ^ keystream[i%16]
	}
	return s.w.Write(encryptedData)
}

// --- 简化版 ChaCha20 实现 (仅教学) ---
type SimpleChaCha20Stream struct {
	r       io.Reader
	w       io.Writer
	key     []byte
	nonce   []byte
	counter uint64
}

func NewSimpleChaCha20Stream(r io.Reader, key, nonce []byte) (*SimpleChaCha20Stream, error) {
	if len(key) != 32 || len(nonce) != 12 {
		return nil, errors.New("invalid key/nonce length for ChaCha20")
	}
	return &SimpleChaCha20Stream{r: r, key: key, nonce: nonce}, nil
}

// 伪加密，实际的ChaCha20 quarter-round 非常复杂
func (s *SimpleChaCha20Stream) pseudoChaChaBlock() []byte {
	// 再次声明，这是占位符
	block := make([]byte, 64)
	copy(block, s.key)
	copy(block[32:], s.nonce)
	binary.BigEndian.PutUint64(block[44:], s.counter)
	s.counter++
	// 用一个简单的hash模拟
	hash := sha256.Sum256(block)
	return hash[:]
}
func (s *SimpleChaCha20Stream) Read(p []byte) (n int, err error) {
	// 简化读取逻辑
	underlyingData := make([]byte, len(p))
	n, err = s.r.Read(underlyingData)
	if n > 0 {
		keystream := s.pseudoChaChaBlock()
		for i := 0; i < n; i++ {
			p[i] = underlyingData[i] ^ keystream[i%32]
		}
	}
	return n, err
}
func (s *SimpleChaCha20Stream) Write(p []byte) (n int, err error) {
	encryptedData := make([]byte, len(p))
	keystream := s.pseudoChaChaBlock()
	for i := 0; i < len(p); i++ {
		encryptedData[i] = p[i] ^ keystream[i%32]
	}
	return s.w.Write(encryptedData)
}
