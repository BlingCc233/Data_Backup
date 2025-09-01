// core/crypto.go
package core

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	// PQC后量子密码学库
	"github.com/cloudflare/circl/kem/kyber/kyber768"   // Kyber KEM
	"github.com/cloudflare/circl/sign/dilithium/mode3" // Dilithium签名
	"golang.org/x/crypto/pbkdf2"
)

// 安全警告: AES和ChaCha20的完整实现仅为教学和理解目的
// 生产环境建议使用经过充分测试的标准库实现

var (
	magicHeader         = []byte("QBAKENCR")
	ErrInvalidMagic     = errors.New("invalid magic header: not an encrypted qbak file")
	ErrPasswordRequired = errors.New("password is required for this encrypted file")
	ErrInvalidPadding   = errors.New("invalid PKCS7 padding")
)

const (
	version        = 0x01
	AlgoAES256_CTR = 0x01
	AlgoChaCha20   = 0x02
	AlgoKyber768   = 0x03 // 后量子KEM
	AlgoDilithium3 = 0x04 // 后量子签名
	AlgoHybrid     = 0x05 // 混合模式：经典+后量子

	// AES常量
	AESBlockSize = 16
	AESKeySize   = 32 // AES-256

	// ChaCha20常量
	ChaChaKeySize   = 32
	ChaNonceSize    = 12
	ChaChaBlockSize = 64
)

// --- 完整AES实现 ---
// (此处省略您已提供的完整AES实现代码... )
var sbox = [256]uint8{
	0x63, 0x7c, 0x77, 0x7b, 0xf2, 0x6b, 0x6f, 0xc5, 0x30, 0x01, 0x67, 0x2b, 0xfe, 0xd7, 0xab, 0x76,
	0xca, 0x82, 0xc9, 0x7d, 0xfa, 0x59, 0x47, 0xf0, 0xad, 0xd4, 0xa2, 0xaf, 0x9c, 0xa4, 0x72, 0xc0,
	0xb7, 0xfd, 0x93, 0x26, 0x36, 0x3f, 0xf7, 0xcc, 0x34, 0xa5, 0xe5, 0xf1, 0x71, 0xd8, 0x31, 0x15,
	0x04, 0xc7, 0x23, 0xc3, 0x18, 0x96, 0x05, 0x9a, 0x07, 0x12, 0x80, 0xe2, 0xeb, 0x27, 0xb2, 0x75,
	0x09, 0x83, 0x2c, 0x1a, 0x1b, 0x6e, 0x5a, 0xa0, 0x52, 0x3b, 0xd6, 0xb3, 0x29, 0xe3, 0x2f, 0x84,
	0x53, 0xd1, 0x00, 0xed, 0x20, 0xfc, 0xb1, 0x5b, 0x6a, 0xcb, 0xbe, 0x39, 0x4a, 0x4c, 0x58, 0xcf,
	0xd0, 0xef, 0xaa, 0xfb, 0x43, 0x4d, 0x33, 0x85, 0x45, 0xf9, 0x02, 0x7f, 0x50, 0x3c, 0x9f, 0xa8,
	0x51, 0xa3, 0x40, 0x8f, 0x92, 0x9d, 0x38, 0xf5, 0xbc, 0xb6, 0xda, 0x21, 0x10, 0xff, 0xf3, 0xd2,
	0xcd, 0x0c, 0x13, 0xec, 0x5f, 0x97, 0x44, 0x17, 0xc4, 0xa7, 0x7e, 0x3d, 0x64, 0x5d, 0x19, 0x73,
	0x60, 0x81, 0x4f, 0xdc, 0x22, 0x2a, 0x90, 0x88, 0x46, 0xee, 0xb8, 0x14, 0xde, 0x5e, 0x0b, 0xdb,
	0xe0, 0x32, 0x3a, 0x0a, 0x49, 0x06, 0x24, 0x5c, 0xc2, 0xd3, 0xac, 0x62, 0x91, 0x95, 0xe4, 0x79,
	0xe7, 0xc8, 0x37, 0x6d, 0x8d, 0xd5, 0x4e, 0xa9, 0x6c, 0x56, 0xf4, 0xea, 0x65, 0x7a, 0xae, 0x08,
	0xba, 0x78, 0x25, 0x2e, 0x1c, 0xa6, 0xb4, 0xc6, 0xe8, 0xdd, 0x74, 0x1f, 0x4b, 0xbd, 0x8b, 0x8a,
	0x70, 0x3e, 0xb5, 0x66, 0x48, 0x03, 0xf6, 0x0e, 0x61, 0x35, 0x57, 0xb9, 0x86, 0xc1, 0x1d, 0x9e,
	0xe1, 0xf8, 0x98, 0x11, 0x69, 0xd9, 0x8e, 0x94, 0x9b, 0x1e, 0x87, 0xe9, 0xce, 0x55, 0x28, 0xdf,
	0x8c, 0xa1, 0x89, 0x0d, 0xbf, 0xe6, 0x42, 0x68, 0x41, 0x99, 0x2d, 0x0f, 0xb0, 0x54, 0xbb, 0x16,
}

var rcon = [11]uint32{
	0x00000000, 0x01000000, 0x02000000, 0x04000000, 0x08000000,
	0x10000000, 0x20000000, 0x40000000, 0x80000000, 0x1b000000, 0x36000000,
}

type AESCipher struct {
	roundKeys [15][4]uint32 // AES-256需要15轮密钥
	rounds    int
}

func NewAESCipher(key []byte) *AESCipher {
	if len(key) != AESKeySize {
		panic("invalid AES key size")
	}

	aes := &AESCipher{rounds: 14} // AES-256有14轮
	aes.keyExpansion(key)
	return aes
}

func (aes *AESCipher) keyExpansion(key []byte) {
	// 将密钥转换为32位字
	w := make([]uint32, 60) // AES-256需要60个字

	for i := 0; i < 8; i++ {
		w[i] = binary.BigEndian.Uint32(key[4*i : 4*i+4])
	}

	for i := 8; i < 60; i++ {
		temp := w[i-1]
		if i%8 == 0 {
			temp = subWord(rotWord(temp)) ^ rcon[i/8]
		} else if i%8 == 4 {
			temp = subWord(temp)
		}
		w[i] = w[i-8] ^ temp
	}

	// 将字转换为轮密钥
	for round := 0; round <= 14; round++ {
		for j := 0; j < 4; j++ {
			aes.roundKeys[round][j] = w[round*4+j]
		}
	}
}

func rotWord(word uint32) uint32 {
	return (word << 8) | (word >> 24)
}

func subWord(word uint32) uint32 {
	return uint32(sbox[(word>>24)&0xff])<<24 |
		uint32(sbox[(word>>16)&0xff])<<16 |
		uint32(sbox[(word>>8)&0xff])<<8 |
		uint32(sbox[word&0xff])
}

func (aes *AESCipher) Encrypt(block []byte) []byte {
	if len(block) != AESBlockSize {
		panic("invalid block size")
	}

	state := make([][]uint8, 4)
	for i := range state {
		state[i] = make([]uint8, 4)
	}

	// 将输入转换为状态矩阵
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state[j][i] = block[i*4+j]
		}
	}

	// 初始轮密钥加
	aes.addRoundKey(state, 0)

	// 主循环
	for round := 1; round < aes.rounds; round++ {
		aes.subBytes(state)
		aes.shiftRows(state)
		aes.mixColumns(state)
		aes.addRoundKey(state, round)
	}

	// 最后一轮（无MixColumns）
	aes.subBytes(state)
	aes.shiftRows(state)
	aes.addRoundKey(state, aes.rounds)

	// 转换回字节数组
	result := make([]byte, AESBlockSize)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			result[i*4+j] = state[j][i]
		}
	}

	return result
}

func (aes *AESCipher) subBytes(state [][]uint8) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			state[i][j] = sbox[state[i][j]]
		}
	}
}

func (aes *AESCipher) shiftRows(state [][]uint8) {
	// 第0行不移位，第1行左移1位，第2行左移2位，第3行左移3位
	temp := make([]uint8, 4)

	for shift := 1; shift < 4; shift++ {
		copy(temp, state[shift])
		for i := 0; i < 4; i++ {
			state[shift][i] = temp[(i+shift)%4]
		}
	}
}

var mixColMatrix = [4][4]uint8{
	{0x02, 0x03, 0x01, 0x01},
	{0x01, 0x02, 0x03, 0x01},
	{0x01, 0x01, 0x02, 0x03},
	{0x03, 0x01, 0x01, 0x02},
}

func (aes *AESCipher) mixColumns(state [][]uint8) {
	for col := 0; col < 4; col++ {
		temp := [4]uint8{}
		for row := 0; row < 4; row++ {
			temp[row] = gmul(mixColMatrix[row][0], state[0][col]) ^
				gmul(mixColMatrix[row][1], state[1][col]) ^
				gmul(mixColMatrix[row][2], state[2][col]) ^
				gmul(mixColMatrix[row][3], state[3][col])
		}
		for row := 0; row < 4; row++ {
			state[row][col] = temp[row]
		}
	}
}

func gmul(a, b uint8) uint8 {
	result := uint8(0)
	for i := 0; i < 8; i++ {
		if (b & 1) != 0 {
			result ^= a
		}
		hiBitSet := (a & 0x80) != 0
		a <<= 1
		if hiBitSet {
			a ^= 0x1b
		}
		b >>= 1
	}
	return result
}

func (aes *AESCipher) addRoundKey(state [][]uint8, round int) {
	for col := 0; col < 4; col++ {
		keyWord := aes.roundKeys[round][col]
		state[0][col] ^= uint8((keyWord >> 24) & 0xff)
		state[1][col] ^= uint8((keyWord >> 16) & 0xff)
		state[2][col] ^= uint8((keyWord >> 8) & 0xff)
		state[3][col] ^= uint8(keyWord & 0xff)
	}
}

// --- AES-CTR模式实现 ---

type AESCTRStream struct {
	r            io.Reader
	w            io.Writer
	cipher       *AESCipher
	iv           []byte
	counter      uint64
	keystream    []byte
	keystreamPos int
}

func NewAESCTRStream(r io.Reader, w io.Writer, key, iv []byte) (*AESCTRStream, error) {
	if len(key) != AESKeySize || len(iv) != AESBlockSize {
		return nil, errors.New("invalid key/iv length for AES-256-CTR")
	}

	// 为了安全性，不直接修改传入的iv
	ivCopy := make([]byte, len(iv))
	copy(ivCopy, iv)

	return &AESCTRStream{
		r:            r,
		w:            w,
		cipher:       NewAESCipher(key),
		iv:           ivCopy,
		counter:      0,
		keystream:    make([]byte, AESBlockSize),
		keystreamPos: AESBlockSize, // 强制第一次生成keystream
	}, nil
}

func (s *AESCTRStream) generateKeystream() {
	counterBlock := make([]byte, AESBlockSize)
	copy(counterBlock, s.iv)

	// 将64位计数器加到IV的后8个字节上，这是CTR模式的常见实现
	// 注意：这会修改counterBlock，但不会修改s.iv
	currentCounterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(currentCounterBytes, s.counter)

	// 将计数器与IV的相应部分相加（带进位）
	var carry uint16
	for i := 7; i >= 0; i-- {
		val := uint16(s.iv[8+i]) + uint16(currentCounterBytes[i]) + carry
		counterBlock[8+i] = byte(val)
		carry = val >> 8
	}

	s.keystream = s.cipher.Encrypt(counterBlock)
	s.counter++
	s.keystreamPos = 0
}

func (s *AESCTRStream) Read(p []byte) (n int, err error) {
	if s.r == nil {
		return 0, errors.New("no reader available")
	}

	// 注意：这里应该是先从底层读取n个字节，然后再解密这n个字节。
	// 您的原始实现可能会导致读取不足的问题。修正如下：
	n, err = s.r.Read(p)
	if n > 0 {
		for i := 0; i < n; i++ {
			if s.keystreamPos >= AESBlockSize {
				s.generateKeystream()
			}
			p[i] ^= s.keystream[s.keystreamPos]
			s.keystreamPos++
		}
	}
	return n, err
}

func (s *AESCTRStream) Write(p []byte) (n int, err error) {
	if s.w == nil {
		return 0, errors.New("no writer available")
	}

	ciphertext := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		if s.keystreamPos >= AESBlockSize {
			s.generateKeystream()
		}
		ciphertext[i] = p[i] ^ s.keystream[s.keystreamPos]
		s.keystreamPos++
	}
	return s.w.Write(ciphertext)
}

// --- ChaCha20完整实现 ---
// (此处省略您已提供的完整ChaCha20实现代码... )
func quarterRound(a, b, c, d *uint32) {
	*a += *b
	*d ^= *a
	*d = (*d << 16) | (*d >> 16)
	*c += *d
	*b ^= *c
	*b = (*b << 12) | (*b >> 20)
	*a += *b
	*d ^= *a
	*d = (*d << 8) | (*d >> 24)
	*c += *d
	*b ^= *c
	*b = (*b << 7) | (*b >> 25)
}

func chaChaBlock(out *[16]uint32, in *[16]uint32) {
	x := *in

	for i := 0; i < 10; i++ {
		// 列轮
		quarterRound(&x[0], &x[4], &x[8], &x[12])
		quarterRound(&x[1], &x[5], &x[9], &x[13])
		quarterRound(&x[2], &x[6], &x[10], &x[14])
		quarterRound(&x[3], &x[7], &x[11], &x[15])
		// 对角轮
		quarterRound(&x[0], &x[5], &x[10], &x[15])
		quarterRound(&x[1], &x[6], &x[11], &x[12])
		quarterRound(&x[2], &x[7], &x[8], &x[13])
		quarterRound(&x[3], &x[4], &x[9], &x[14])
	}

	for i := 0; i < 16; i++ {
		out[i] = x[i] + in[i]
	}
}

type ChaCha20Stream struct {
	r            io.Reader
	w            io.Writer
	key          [8]uint32
	nonce        [3]uint32
	counter      uint32
	keystream    []byte
	keystreamPos int
}

func NewChaCha20Stream(r io.Reader, w io.Writer, key, nonce []byte) (*ChaCha20Stream, error) {
	if len(key) != ChaChaKeySize || len(nonce) != ChaNonceSize {
		return nil, errors.New("invalid key/nonce length for ChaCha20")
	}

	s := &ChaCha20Stream{
		r:            r,
		w:            w,
		keystream:    make([]byte, ChaChaBlockSize),
		keystreamPos: ChaChaBlockSize, // 强制第一次生成
	}

	// 设置密钥
	for i := 0; i < 8; i++ {
		s.key[i] = binary.LittleEndian.Uint32(key[i*4 : i*4+4])
	}

	// 设置nonce
	for i := 0; i < 3; i++ {
		s.nonce[i] = binary.LittleEndian.Uint32(nonce[i*4 : i*4+4])
	}

	return s, nil
}

func (s *ChaCha20Stream) generateKeystream() {
	var state, out [16]uint32

	// ChaCha20常量 "expand 32-byte k"
	state[0] = 0x61707865
	state[1] = 0x3320646e
	state[2] = 0x79622d32
	state[3] = 0x6b206574

	// 密钥
	copy(state[4:12], s.key[:])

	// 计数器和nonce
	state[12] = s.counter
	copy(state[13:16], s.nonce[:])

	chaChaBlock(&out, &state)

	// 转换为字节
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(s.keystream[i*4:i*4+4], out[i])
	}

	s.counter++
	s.keystreamPos = 0
}

func (s *ChaCha20Stream) Read(p []byte) (n int, err error) {
	if s.r == nil {
		return 0, errors.New("no reader available")
	}

	// 修正同AESCTRStream.Read
	n, err = s.r.Read(p)
	if n > 0 {
		for i := 0; i < n; i++ {
			if s.keystreamPos >= ChaChaBlockSize {
				s.generateKeystream()
			}
			p[i] ^= s.keystream[s.keystreamPos]
			s.keystreamPos++
		}
	}
	return n, err
}

func (s *ChaCha20Stream) Write(p []byte) (n int, err error) {
	if s.w == nil {
		return 0, errors.New("no writer available")
	}

	ciphertext := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		if s.keystreamPos >= ChaChaBlockSize {
			s.generateKeystream()
		}
		ciphertext[i] = p[i] ^ s.keystream[s.keystreamPos]
		s.keystreamPos++
	}
	return s.w.Write(ciphertext)
}

// --- 后量子密码学(PQC)实现 ---

// Kyber KEM包装器
type KyberKEM struct {
	publicKey  kyber768.PublicKey
	privateKey kyber768.PrivateKey
}

func GenerateKyberKeyPair() (*KyberKEM, error) {
	pub, priv, err := kyber768.GenerateKeyPair(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &KyberKEM{
		publicKey:  *pub,
		privateKey: *priv,
	}, nil
}

func (k *KyberKEM) Encapsulate() (ciphertext, sharedSecret []byte, err error) {
	// 接收方公钥加密
	return kyber768.Scheme().Encapsulate(&k.publicKey)
}

func (k *KyberKEM) Decapsulate(ciphertext []byte) (sharedSecret []byte, err error) {
	// 接收方私钥解密
	return kyber768.Scheme().Decapsulate(&k.privateKey, ciphertext)
}

// Dilithium签名包装器
type DilithiumSigner struct {
	publicKey  mode3.PublicKey
	privateKey mode3.PrivateKey
}

func GenerateDilithiumKeyPair() (*DilithiumSigner, error) {
	pub, priv, err := mode3.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &DilithiumSigner{
		publicKey:  *pub,
		privateKey: *priv,
	}, nil
}

func (d *DilithiumSigner) Sign(message []byte) error {
	sigbytes := []byte("blingcc")
	mode3.SignTo(&d.privateKey, message, sigbytes)
	// 使用私钥签名
	return nil
}

func (d *DilithiumSigner) Verify(message, signature []byte) bool {
	// 使用公钥验签
	return mode3.Verify(&d.publicKey, message, signature)
}

// --- 混合模式加密(经典+后量子) ---

type HybridEncryption struct {
	kyberKEM *KyberKEM
}

func NewHybridEncryption() (*HybridEncryption, error) {
	kyberKEM, err := GenerateKyberKeyPair()
	if err != nil {
		return nil, err
	}

	return &HybridEncryption{
		kyberKEM: kyberKEM,
	}, nil
}

// --- 密钥派生函数 ---
func deriveKey(password string, salt []byte) []byte {
	return pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New)
}

// --- 加密写入器 ---
type encryptedWriter struct {
	underlyingWriter io.Writer
	streamCipher     io.Writer
	// io.Closer for future use if needed (e.g., for file handles)
}

// NewEncryptedWriterForHybrid is a helper for testing hybrid mode.
// It returns the writer and the private key needed for decryption.
func NewEncryptedWriterForHybrid(w io.Writer, algorithm uint8) (io.WriteCloser, *kyber768.PrivateKey, error) {
	if algorithm != AlgoHybrid {
		return nil, nil, fmt.Errorf("this function is only for hybrid mode")
	}

	hybrid, err := NewHybridEncryption()
	if err != nil {
		return nil, nil, err
	}

	ciphertext, sharedSecret, err := hybrid.kyberKEM.Encapsulate()
	if err != nil {
		return nil, nil, err
	}

	aesIV := make([]byte, AESBlockSize)
	if _, err := rand.Read(aesIV); err != nil {
		return nil, nil, err
	}

	nonce := append(ciphertext, aesIV...)

	aesStream, err := NewAESCTRStream(nil, w, sharedSecret, aesIV)
	if err != nil {
		return nil, nil, err
	}

	header := new(bytes.Buffer)
	header.Write(magicHeader)
	header.WriteByte(version)
	header.WriteByte(algorithm)
	// Hybrid mode doesn't use password-derived salt, so salt length is 0
	header.WriteByte(0)
	header.WriteByte(byte(len(nonce)))
	header.Write(nonce)

	if _, err := w.Write(header.Bytes()); err != nil {
		return nil, nil, err
	}

	writer := &encryptedWriter{
		underlyingWriter: w,
		streamCipher:     aesStream,
	}
	return writer, &hybrid.kyberKEM.privateKey, nil
}

func NewEncryptedWriter(w io.Writer, password string, algorithm uint8) (io.WriteCloser, error) {
	if algorithm != AlgoHybrid && password == "" {
		return nil, errors.New("password cannot be empty for encryption")
	}

	// AlgoHybrid has a different key management scheme (KEM) and doesn't use a password.
	if algorithm == AlgoHybrid {
		// NOTE: In a real application, you would use the recipient's public key here.
		// For this example, we generate a new key pair and discard the private key.
		// Decryption would require a different function that takes the private key.
		// See NewEncryptedWriterForHybrid for a testable implementation.
		return nil, fmt.Errorf("hybrid encryption requires recipient's public key; use test helper function instead")
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
		nonce = make([]byte, AESBlockSize)
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}
		aesStream, err := NewAESCTRStream(nil, w, key, nonce)
		if err != nil {
			return nil, err
		}
		stream = aesStream

	case AlgoChaCha20:
		nonce = make([]byte, ChaNonceSize)
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}
		chachaStream, err := NewChaCha20Stream(nil, w, key, nonce)
		if err != nil {
			return nil, err
		}
		stream = chachaStream
	default:
		return nil, fmt.Errorf("unsupported algorithm: %d", algorithm)
	}

	// 写入文件头
	header := new(bytes.Buffer)
	header.Write(magicHeader)
	header.WriteByte(version)
	header.WriteByte(algorithm)
	header.WriteByte(byte(len(salt)))
	header.Write(salt)
	header.WriteByte(byte(len(nonce)))
	header.Write(nonce)

	if _, err := w.Write(header.Bytes()); err != nil {
		return nil, err
	}

	return &encryptedWriter{
		underlyingWriter: w,
		streamCipher:     stream,
	}, nil
}

func (ew *encryptedWriter) Write(p []byte) (n int, err error) {
	return ew.streamCipher.Write(p)
}

func (ew *encryptedWriter) Close() error {
	// In this implementation, the underlying writer is managed by the caller.
	// If this struct were to own a file handle, this is where it would be closed.
	return nil
}

// --- 解密读取器 ---

type decryptedReader struct {
	streamCipher io.Reader
}

func (dr *decryptedReader) Read(p []byte) (n int, err error) {
	return dr.streamCipher.Read(p)
}

// NewDecryptedReaderForHybrid is a helper for testing hybrid mode.
// It requires the private key that corresponds to the public key used for encryption.
func NewDecryptedReaderForHybrid(r io.Reader, privKey *kyber768.PrivateKey) (io.Reader, error) {
	// Header validation (similar to NewDecryptedReader)
	header := make([]byte, len(magicHeader))
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	if !bytes.Equal(header, magicHeader) {
		return nil, ErrInvalidMagic
	}
	meta := make([]byte, 2) // version, algo
	if _, err := io.ReadFull(r, meta); err != nil {
		return nil, err
	}
	if meta[1] != AlgoHybrid {
		return nil, fmt.Errorf("expected hybrid algorithm, got %d", meta[1])
	}

	// Read salt (length is 0 for hybrid)
	saltLenBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, saltLenBuf); err != nil {
		return nil, err
	}

	// Read nonce (which is KyberCT + AES_IV)
	nonceLenBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, nonceLenBuf); err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceLenBuf[0])
	if _, err := io.ReadFull(r, nonce); err != nil {
		return nil, err
	}

	// Decapsulate to get the shared secret
	kyberCT := nonce[:kyber768.CiphertextSize]
	aesIV := nonce[kyber768.CiphertextSize:]

	sharedSecret, err := kyber768.Scheme().Decapsulate(privKey, kyberCT)
	if err != nil {
		return nil, fmt.Errorf("failed to decapsulate kyber ciphertext: %w", err)
	}

	aesStream, err := NewAESCTRStream(r, nil, sharedSecret, aesIV)
	if err != nil {
		return nil, err
	}

	return &decryptedReader{streamCipher: aesStream}, nil
}

func NewDecryptedReader(r io.Reader, password string) (io.Reader, error) {
	// 读取魔数
	header := make([]byte, len(magicHeader))
	if _, err := io.ReadFull(r, header); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, ErrInvalidMagic
		}
		return nil, err
	}
	if !bytes.Equal(header, magicHeader) {
		return nil, ErrInvalidMagic
	}

	// 读取版本和算法
	meta := make([]byte, 2)
	if _, err := io.ReadFull(r, meta); err != nil {
		return nil, err
	}
	// versionByte := meta[0]
	algoByte := meta[1]

	// 混合模式有不同的处理流程，因为它不使用密码
	if algoByte == AlgoHybrid {
		return nil, fmt.Errorf("hybrid encrypted files require the recipient's private key for decryption, use NewDecryptedReaderForHybrid helper")
	}

	if password == "" {
		return nil, ErrPasswordRequired
	}

	// 读取Salt
	saltLenBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, saltLenBuf); err != nil {
		return nil, err
	}
	salt := make([]byte, saltLenBuf[0])
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, err
	}

	// 读取Nonce/IV
	nonceLenBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, nonceLenBuf); err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceLenBuf[0])
	if _, err := io.ReadFull(r, nonce); err != nil {
		return nil, err
	}

	// 派生密钥
	key := deriveKey(password, salt)

	var stream io.Reader
	var err error

	switch algoByte {
	case AlgoAES256_CTR:
		stream, err = NewAESCTRStream(r, nil, key, nonce)
		if err != nil {
			return nil, err
		}
	case AlgoChaCha20:
		stream, err = NewChaCha20Stream(r, nil, key, nonce)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported or invalid algorithm: %d", algoByte)
	}

	return &decryptedReader{streamCipher: stream}, nil
}
