// core/crypto.go
package core

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"runtime"
	"sync"
)

// --- CC的 SHA-256 实现 ---
// 该实现遵循 FIPS 180-4 标准

const (
	sha256Size      = 32 // SHA256校验和的字节大小
	sha256BlockSize = 64 //  SHA256哈希算法的块大小
)

// SHA-256 的轮常量
var K = [64]uint32{
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
}

// SHA-256 的初始哈希值
var H0 = [8]uint32{
	0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
}

type digest struct {
	h   [8]uint32
	x   [sha256BlockSize]byte
	nx  int
	len uint64
}

func (d *digest) Reset() {
	d.h = H0
	d.nx = 0
	d.len = 0
}

func New() *digest {
	d := new(digest)
	d.Reset()
	return d
}

func (d *digest) Size() int { return sha256Size }

func (d *digest) BlockSize() int { return sha256BlockSize }

func (d *digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	d.len += uint64(nn)
	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == sha256BlockSize {
			block(d, d.x[:])
			d.nx = 0
		}
		p = p[n:]
	}
	if len(p) >= sha256BlockSize {
		n := len(p) &^ (sha256BlockSize - 1)
		block(d, p[:n])
		p = p[n:]
	}
	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}
	return
}

func (d *digest) Sum(in []byte) []byte {
	d0 := *d
	hash := d0.checkSum()
	return append(in, hash[:]...)
}

func (d *digest) checkSum() [sha256Size]byte {
	len := d.len
	// 填充: 添加 1 bit
	var tmp [sha256BlockSize]byte
	tmp[0] = 0x80
	if len%sha256BlockSize < 56 {
		d.Write(tmp[0 : 56-len%sha256BlockSize])
	} else {
		d.Write(tmp[0 : sha256BlockSize+56-len%sha256BlockSize])
	}

	// 填充: 添加长度
	binary.BigEndian.PutUint64(tmp[:], len<<3)
	d.Write(tmp[0:8])

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [sha256Size]byte
	for i, s := range d.h {
		binary.BigEndian.PutUint32(digest[i*4:], s)
	}
	return digest
}

// Sum256 计算数据的 SHA256 校验和
func Sum256(data []byte) [sha256Size]byte {
	d := New()
	d.Write(data)
	return d.checkSum()
}

func block(dig *digest, p []byte) {
	var w [64]uint32
	h := dig.h
	for len(p) >= sha256BlockSize {
		// 初始化消息调度数组
		for i := 0; i < 16; i++ {
			w[i] = binary.BigEndian.Uint32(p[i*4:])
		}
		for i := 16; i < 64; i++ {
			s0 := rotateRight32(w[i-15], 7) ^ rotateRight32(w[i-15], 18) ^ (w[i-15] >> 3)
			s1 := rotateRight32(w[i-2], 17) ^ rotateRight32(w[i-2], 19) ^ (w[i-2] >> 10)
			w[i] = w[i-16] + s0 + w[i-7] + s1
		}

		// 初始化工作变量
		a, b, c, d, e, f, g, h_ := h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7]

		// 压缩循环
		for i := 0; i < 64; i++ {
			s1 := rotateRight32(e, 6) ^ rotateRight32(e, 11) ^ rotateRight32(e, 25)
			ch := (e & f) ^ (^e & g)
			temp1 := h_ + s1 + ch + K[i] + w[i]
			s0 := rotateRight32(a, 2) ^ rotateRight32(a, 13) ^ rotateRight32(a, 22)
			maj := (a & b) ^ (a & c) ^ (b & c)
			temp2 := s0 + maj

			h_ = g
			g = f
			f = e
			e = d + temp1
			d = c
			c = b
			b = a
			a = temp1 + temp2
		}

		// 更新哈希值
		h[0] += a
		h[1] += b
		h[2] += c
		h[3] += d
		h[4] += e
		h[5] += f
		h[6] += g
		h[7] += h_

		p = p[sha256BlockSize:]
	}
	dig.h = h
}

// --- 加密 ---

var (
	magicHeader         = []byte("QBAKENCR")
	ErrInvalidMagic     = errors.New("invalid magic header: not an encrypted qbak file")
	ErrPasswordRequired = errors.New("password is required for this encrypted file")
	ErrInvalidKeySize   = errors.New("invalid key size")
	ErrInvalidNonceSize = errors.New("invalid nonce size")
)

const (
	version1       = 0x01
	version2       = 0x02
	currentVersion = version2
	AlgoAES256_CTR = 0x01
	AlgoChaCha20   = 0x02
)

// --- 密钥派生 (PBKDF2 CC的实现) ---

func pbkdf2(password, salt []byte, iter, keyLen int) []byte {
	hashLen := sha256Size
	l := (keyLen + hashLen - 1) / hashLen
	r := keyLen - (l-1)*hashLen

	result := make([]byte, 0, l*hashLen)

	for i := 1; i <= l; i++ {
		ui := prf(password, append(salt, byte(i>>24), byte(i>>16), byte(i>>8), byte(i)))
		u := make([]byte, len(ui))
		copy(u, ui)

		for j := 1; j < iter; j++ {
			ui = prf(password, ui)
			for k := range u {
				u[k] ^= ui[k]
			}
		}
		result = append(result, u...)
	}

	if r < hashLen {
		return result[:keyLen]
	}
	return result[:keyLen]
}

func prf(key, data []byte) []byte {
	// HMAC-SHA256 实现
	blockSize := sha256BlockSize
	if len(key) > blockSize {
		h := Sum256(key)
		key = h[:]
	}

	if len(key) < blockSize {
		newKey := make([]byte, blockSize)
		copy(newKey, key)
		key = newKey
	}

	opad := make([]byte, blockSize)
	ipad := make([]byte, blockSize)

	for i := range key {
		opad[i] = key[i] ^ 0x5c
		ipad[i] = key[i] ^ 0x36
	}

	inner := New()
	inner.Write(ipad)
	inner.Write(data)
	innerHash := inner.Sum(nil)

	outer := New()
	outer.Write(opad)
	outer.Write(innerHash)
	return outer.Sum(nil)
}

func deriveKey(password string, salt []byte) []byte {
	return pbkdf2([]byte(password), salt, 4096, 32)
}

// --- AES-256 CC的实现 ---

var sbox = [256]byte{
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

type AES struct {
	roundKeys [15][4]uint32 // 最多14轮加1轮初始
	rounds    int
}

func NewAES(key []byte) (*AES, error) {
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, ErrInvalidKeySize
	}

	aes := &AES{}
	switch keyLen {
	case 16:
		aes.rounds = 10
	case 24:
		aes.rounds = 12
	case 32:
		aes.rounds = 14
	}

	aes.keyExpansion(key)
	return aes, nil
}

func (aes *AES) keyExpansion(key []byte) {
	keyWords := len(key) / 4

	// 初始化前N个字
	for i := 0; i < keyWords; i++ {
		aes.roundKeys[i/4][i%4] = binary.BigEndian.Uint32(key[i*4 : (i+1)*4])
	}

	// 生成剩余字
	for i := keyWords; i < 4*(aes.rounds+1); i++ {
		temp := aes.roundKeys[(i-1)/4][(i-1)%4]

		if i%keyWords == 0 {
			temp = aes.subWord(aes.rotWord(temp)) ^ rcon[i/keyWords]
		} else if keyWords > 6 && i%keyWords == 4 {
			temp = aes.subWord(temp)
		}

		aes.roundKeys[i/4][i%4] = aes.roundKeys[(i-keyWords)/4][(i-keyWords)%4] ^ temp
	}
}

func (aes *AES) subWord(word uint32) uint32 {
	return uint32(sbox[word>>24])<<24 |
		uint32(sbox[(word>>16)&0xff])<<16 |
		uint32(sbox[(word>>8)&0xff])<<8 |
		uint32(sbox[word&0xff])
}

func (aes *AES) rotWord(word uint32) uint32 {
	return word<<8 | word>>24
}

func (aes *AES) Encrypt(plaintext []byte) []byte {
	if len(plaintext) != 16 {
		panic("AES block size must be 16 bytes")
	}

	state := make([][]byte, 4)
	for i := range state {
		state[i] = make([]byte, 4)
		for j := range state[i] {
			// AES state is column-major: state[row][col] = in[4*col+row]
			state[i][j] = plaintext[j*4+i]
		}
	}

	aes.addRoundKey(state, 0)

	for round := 1; round < aes.rounds; round++ {
		aes.subBytes(state)
		aes.shiftRows(state)
		aes.mixColumns(state)
		aes.addRoundKey(state, round)
	}

	aes.subBytes(state)
	aes.shiftRows(state)
	aes.addRoundKey(state, aes.rounds)

	ciphertext := make([]byte, 16)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			ciphertext[j*4+i] = state[i][j]
		}
	}

	return ciphertext
}

func (aes *AES) subBytes(state [][]byte) {
	for i := range state {
		for j := range state[i] {
			state[i][j] = sbox[state[i][j]]
		}
	}
}

func (aes *AES) shiftRows(state [][]byte) {
	temp := make([]byte, 4)
	for i := 1; i < 4; i++ {
		copy(temp, state[i])
		for j := 0; j < 4; j++ {
			state[i][j] = temp[(j+i)%4]
		}
	}
}

func (aes *AES) mixColumns(state [][]byte) {
	for c := 0; c < 4; c++ {
		a := make([]byte, 4)
		copy(a, []byte{state[0][c], state[1][c], state[2][c], state[3][c]})

		state[0][c] = aes.galoisMul(a[0], 0x02) ^ aes.galoisMul(a[1], 0x03) ^ a[2] ^ a[3]
		state[1][c] = a[0] ^ aes.galoisMul(a[1], 0x02) ^ aes.galoisMul(a[2], 0x03) ^ a[3]
		state[2][c] = a[0] ^ a[1] ^ aes.galoisMul(a[2], 0x02) ^ aes.galoisMul(a[3], 0x03)
		state[3][c] = aes.galoisMul(a[0], 0x03) ^ a[1] ^ a[2] ^ aes.galoisMul(a[3], 0x02)
	}
}

func (aes *AES) galoisMul(a, b byte) byte {
	p := byte(0)
	for i := 0; i < 8; i++ {
		if b&1 != 0 {
			p ^= a
		}
		carry := a & 0x80
		a <<= 1
		if carry != 0 {
			a ^= 0x1b
		}
		b >>= 1
	}
	return p
}

func (aes *AES) addRoundKey(state [][]byte, round int) {
	for i := 0; i < 4; i++ {
		key := aes.roundKeys[round][i]
		state[0][i] ^= byte(key >> 24)
		state[1][i] ^= byte(key >> 16)
		state[2][i] ^= byte(key >> 8)
		state[3][i] ^= byte(key)
	}
}

// --- ChaCha20 CC的实现 ---

type ChaCha20 struct {
	key     [8]uint32
	counter uint32
	nonce   [3]uint32
}

func NewChaCha20(key []byte, nonce []byte) (*ChaCha20, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}
	if len(nonce) != 12 {
		return nil, ErrInvalidNonceSize
	}

	c := &ChaCha20{}

	for i := 0; i < 8; i++ {
		c.key[i] = binary.LittleEndian.Uint32(key[i*4 : (i+1)*4])
	}

	for i := 0; i < 3; i++ {
		c.nonce[i] = binary.LittleEndian.Uint32(nonce[i*4 : (i+1)*4])
	}

	return c, nil
}

func (c *ChaCha20) quarterRound(a, b, c_idx, d *uint32) {
	*a += *b
	*d ^= *a
	*d = bits.RotateLeft32(*d, 16)

	*c_idx += *d
	*b ^= *c_idx
	*b = bits.RotateLeft32(*b, 12)

	*a += *b
	*d ^= *a
	*d = bits.RotateLeft32(*d, 8)

	*c_idx += *d
	*b ^= *c_idx
	*b = bits.RotateLeft32(*b, 7)
}

func (c *ChaCha20) block() [64]byte {
	// ChaCha20 状态
	state := [16]uint32{
		0x61707865, 0x3320646e, 0x79622d32, 0x6b206574, // "expand 32-byte k"
		c.key[0], c.key[1], c.key[2], c.key[3],
		c.key[4], c.key[5], c.key[6], c.key[7],
		c.counter, c.nonce[0], c.nonce[1], c.nonce[2],
	}

	working := state

	// 20轮 (10次双轮)
	for i := 0; i < 10; i++ {
		// 列轮
		c.quarterRound(&working[0], &working[4], &working[8], &working[12])
		c.quarterRound(&working[1], &working[5], &working[9], &working[13])
		c.quarterRound(&working[2], &working[6], &working[10], &working[14])
		c.quarterRound(&working[3], &working[7], &working[11], &working[15])

		// 对角轮
		c.quarterRound(&working[0], &working[5], &working[10], &working[15])
		c.quarterRound(&working[1], &working[6], &working[11], &working[12])
		c.quarterRound(&working[2], &working[7], &working[8], &working[13])
		c.quarterRound(&working[3], &working[4], &working[9], &working[14])
	}

	// 加上初始状态
	for i := range working {
		working[i] += state[i]
	}

	// 转换为字节序列
	var result [64]byte
	for i, word := range working {
		binary.LittleEndian.PutUint32(result[i*4:(i+1)*4], word)
	}

	c.counter++
	return result
}

// --- 流密码实现 ---

type CipherStream interface {
	XORKeyStream(dst, src []byte)
	// SetCounter 允许将密码流的状态重置到处理特定块的位置
	SetCounter(counter uint64)
}

type AESCTRStream struct {
	aes       *AES
	iv        [16]byte
	counter   uint64
	keystream []byte
	pos       int
}

func NewAESCTRStream(key, iv []byte) (*AESCTRStream, error) {
	if len(iv) != 16 {
		return nil, ErrInvalidNonceSize
	}

	aes, err := NewAES(key)
	if err != nil {
		return nil, err
	}

	stream := &AESCTRStream{aes: aes}
	copy(stream.iv[:], iv)

	return stream, nil
}

func (s *AESCTRStream) SetCounter(counter uint64) {
	s.counter = counter
	s.pos = 16 // Force regeneration of keystream
}

func (s *AESCTRStream) XORKeyStream(dst, src []byte) {
	for len(src) > 0 {
		if s.pos >= len(s.keystream) {
			s.generateKeystream()
		}

		n := len(s.keystream) - s.pos
		if n > len(src) {
			n = len(src)
		}

		// XOR a portion of the keystream
		for i := 0; i < n; i++ {
			dst[i] = src[i] ^ s.keystream[s.pos+i]
		}

		src = src[n:]
		dst = dst[n:]
		s.pos += n
	}
}

func (s *AESCTRStream) generateKeystream() {
	counterBlock := s.iv
	binary.BigEndian.PutUint64(counterBlock[8:], s.counter)
	s.keystream = s.aes.Encrypt(counterBlock[:])
	s.counter++
	s.pos = 0
}

type ChaCha20Stream struct {
	chacha    *ChaCha20
	keystream []byte
	pos       int
}

func NewChaCha20Stream(key, nonce []byte) (*ChaCha20Stream, error) {
	chacha, err := NewChaCha20(key, nonce)
	if err != nil {
		return nil, err
	}

	return &ChaCha20Stream{chacha: chacha}, nil
}

func (s *ChaCha20Stream) SetCounter(counter uint64) {
	s.chacha.counter = uint32(counter) // ChaCha20 counter is u32
	s.pos = 64                         // Force regeneration
}

func (s *ChaCha20Stream) XORKeyStream(dst, src []byte) {
	for len(src) > 0 {
		if s.pos >= len(s.keystream) {
			s.generateKeystream()
		}

		n := len(src)
		if n > len(s.keystream)-s.pos {
			n = len(s.keystream) - s.pos
		}

		for i := 0; i < n; i++ {
			dst[i] = src[i] ^ s.keystream[s.pos+i]
		}

		src = src[n:]
		dst = dst[n:]
		s.pos += n
	}
}

func (s *ChaCha20Stream) generateKeystream() {
	block := s.chacha.block()
	s.keystream = block[:]
	s.pos = 0
}

// --- 后量子密码学支持 ---
//
// TODO
//

// --- 并行流读写器实现 (完全重写和修复) ---

const (
	chunkSize = 1 * 1024 * 1024
)

var (
	workerCount = runtime.NumCPU()
)

type job struct {
	id   int
	data []byte
}

type result struct {
	id   int
	data []byte
}

// parallelStreamWriter 是一个并发安全的 io.WriteCloser，用于并行加密数据。
type parallelStreamWriter struct {
	w             io.Writer
	newStreamFn   func() (CipherStream, error)
	blockPerChunk uint64

	jobs    chan job
	results chan result

	wg           sync.WaitGroup // 用于等待 workers
	aggregatorWg sync.WaitGroup // 用于等待 aggregator

	buffer []byte
	nextID int

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	writeErr  error
	errLock   sync.Mutex
}

func newParallelStreamWriter(w io.Writer, algorithm uint8, key, nonce []byte) (*parallelStreamWriter, error) {
	var newStreamFn func() (CipherStream, error)
	var streamBlockSize int

	switch algorithm {
	case AlgoAES256_CTR:
		newStreamFn = func() (CipherStream, error) { return NewAESCTRStream(key, nonce) }
		streamBlockSize = 16
	case AlgoChaCha20:
		newStreamFn = func() (CipherStream, error) { return NewChaCha20Stream(key, nonce) }
		streamBlockSize = 64
	default:
		return nil, fmt.Errorf("unsupported algorithm: %d", algorithm)
	}

	ctx, cancel := context.WithCancel(context.Background())

	sw := &parallelStreamWriter{
		w:             w,
		newStreamFn:   newStreamFn,
		blockPerChunk: uint64(chunkSize / streamBlockSize),
		jobs:          make(chan job, workerCount),
		results:       make(chan result, workerCount),
		buffer:        make([]byte, 0, chunkSize),
		ctx:           ctx,
		cancel:        cancel,
	}

	sw.start()
	return sw, nil
}

func (sw *parallelStreamWriter) start() {
	// 启动 workers
	sw.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go sw.worker()
	}

	// 启动 aggregator
	sw.aggregatorWg.Add(1)
	go sw.aggregator()
}

func (sw *parallelStreamWriter) worker() {
	defer sw.wg.Done()
	stream, err := sw.newStreamFn()
	if err != nil {
		sw.setErr(fmt.Errorf("worker failed to create stream: %w", err))
		return
	}

	for {
		select {
		case <-sw.ctx.Done():
			return
		case j, ok := <-sw.jobs:
			if !ok {
				return // jobs channel closed, normal exit
			}

			counter := uint64(j.id) * sw.blockPerChunk
			stream.SetCounter(counter)

			encrypted := make([]byte, len(j.data))
			stream.XORKeyStream(encrypted, j.data)

			select {
			case sw.results <- result{id: j.id, data: encrypted}:
			case <-sw.ctx.Done():
				return // Abort sending if context is cancelled
			}
		}
	}
}

func (sw *parallelStreamWriter) aggregator() {
	defer sw.aggregatorWg.Done()

	resultsBuffer := make(map[int][]byte)
	nextWriteID := 0

	for {
		select {
		case <-sw.ctx.Done():
			return // Exit if context is cancelled
		case res, ok := <-sw.results:
			if !ok {
				return // results channel closed, normal exit
			}

			resultsBuffer[res.id] = res.data

			// 循环写入所有已按顺序到达的块
			for {
				data, exists := resultsBuffer[nextWriteID]
				if !exists {
					break // 下一个块还没到
				}

				if _, err := sw.w.Write(data); err != nil {
					sw.setErr(fmt.Errorf("aggregator failed to write chunk %d: %w", nextWriteID, err))
					return // 发生写入错误，停止聚合
				}

				delete(resultsBuffer, nextWriteID)
				nextWriteID++
			}
		}
	}
}

func (sw *parallelStreamWriter) Write(p []byte) (n int, err error) {
	if err := sw.getErr(); err != nil {
		return 0, err
	}

	sw.buffer = append(sw.buffer, p...)

	for len(sw.buffer) >= chunkSize {
		chunk := make([]byte, chunkSize)
		copy(chunk, sw.buffer[:chunkSize])

		select {
		case sw.jobs <- job{id: sw.nextID, data: chunk}:
			sw.nextID++
			sw.buffer = sw.buffer[chunkSize:]
		case <-sw.ctx.Done():
			return 0, sw.getErr() // Return error if pipeline has been cancelled
		}
	}

	return len(p), nil
}

func (sw *parallelStreamWriter) Close() error {
	sw.closeOnce.Do(func() {
		// 发送剩余的 buffer 数据
		if len(sw.buffer) > 0 && sw.getErr() == nil {
			sw.jobs <- job{id: sw.nextID, data: sw.buffer}
			sw.buffer = nil
		}

		// 关闭 jobs channel，通知 workers 不再有新任务
		close(sw.jobs)

		// 等待所有 workers 完成
		sw.wg.Wait()

		// 关闭 results channel，通知 aggregator 不再有新结果
		close(sw.results)

		// 等待 aggregator 完成
		sw.aggregatorWg.Wait()

		// 停止所有协程，以防万一
		sw.cancel()

		// 关闭底层的 writer (如果需要)
		if c, ok := sw.w.(io.Closer); ok {
			if err := c.Close(); err != nil {
				sw.setErr(err)
			}
		}
	})

	return sw.getErr()
}

func (sw *parallelStreamWriter) setErr(err error) {
	sw.errLock.Lock()
	if sw.writeErr == nil {
		sw.writeErr = err
		sw.cancel() // 发生错误时，取消 context
	}
	sw.errLock.Unlock()
}

func (sw *parallelStreamWriter) getErr() error {
	sw.errLock.Lock()
	defer sw.errLock.Unlock()
	return sw.writeErr
}

// newParallelStreamReaderWithPipe 使用 io.Pipe 创建一个并行的解密 io.Reader
func newParallelStreamReaderWithPipe(r io.Reader, algorithm uint8, key, nonce []byte) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	var newStreamFn func() (CipherStream, error)
	var streamBlockSize int

	switch algorithm {
	case AlgoAES256_CTR:
		newStreamFn = func() (CipherStream, error) { return NewAESCTRStream(key, nonce) }
		streamBlockSize = 16
	case AlgoChaCha20:
		newStreamFn = func() (CipherStream, error) { return NewChaCha20Stream(key, nonce) }
		streamBlockSize = 64
	default:
		return nil, fmt.Errorf("unsupported algorithm: %d", algorithm)
	}

	blockPerChunk := uint64(chunkSize / streamBlockSize)

	go func() {
		// 在流水线结束时，将错误传递给 reader
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

		jobs := make(chan job, workerCount)
		results := make(chan result, workerCount)
		var wg sync.WaitGroup

		// 启动 workers
		wg.Add(workerCount)
		for i := 0; i < workerCount; i++ {
			go func() {
				defer wg.Done()
				stream, err := newStreamFn()
				if err != nil {
					pipelineErr = fmt.Errorf("worker stream init failed: %w", err)
					cancel()
					return
				}
				for {
					select {
					case <-ctx.Done():
						return
					case j, ok := <-jobs:
						if !ok {
							return
						}

						counter := uint64(j.id) * blockPerChunk
						stream.SetCounter(counter)
						decrypted := make([]byte, len(j.data))
						stream.XORKeyStream(decrypted, j.data)

						select {
						case results <- result{id: j.id, data: decrypted}:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
		}

		// 启动 producer (读取加密文件并分发任务)
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

				buf := make([]byte, chunkSize)
				n, err := io.ReadFull(r, buf)
				if n > 0 {
					jobs <- job{id: nextID, data: buf[:n]}
					nextID++
				}
				if err != nil {
					if err != io.EOF && err != io.ErrUnexpectedEOF {
						pipelineErr = err
						cancel()
					}
					return
				}
			}
		}()

		// 启动一个协程等待所有任务完成并关闭 results channel
		go func() {
			producerWg.Wait()
			wg.Wait()
			close(results)
		}()

		// Aggregator (重组解密后的数据并写入 pipe)
		resultsBuffer := make(map[int][]byte)
		nextWriteID := 0

		for {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-results:
				if !ok {
					return
				} // 正常结束

				resultsBuffer[res.id] = res.data
				for {
					data, exists := resultsBuffer[nextWriteID]
					if !exists {
						break
					}

					if _, err := pw.Write(data); err != nil {
						// Pipe 的 reader 端已关闭，说明消费者不再需要数据，
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

// --- 流读写器实现 ---

type streamWriter struct {
	w      io.Writer
	stream interface{ XORKeyStream(dst, src []byte) }
	buf    []byte
}

func (sw *streamWriter) Write(p []byte) (n int, err error) {
	encrypted := make([]byte, len(p))
	sw.stream.XORKeyStream(encrypted, p)
	return sw.w.Write(encrypted)
}

func (sw *streamWriter) Close() error {
	if c, ok := sw.w.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

type streamReader struct {
	r      io.Reader
	stream interface{ XORKeyStream(dst, src []byte) }
}

func (sr *streamReader) Read(p []byte) (n int, err error) {
	n, err = sr.r.Read(p)
	if n > 0 {
		decrypted := make([]byte, n)
		sr.stream.XORKeyStream(decrypted, p[:n])
		copy(p[:n], decrypted)
	}
	return n, err
}

// --- 加密写入器 ---

func NewEncryptedWriter(w io.Writer, password string, algorithm uint8) (io.WriteCloser, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty for encryption")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := deriveKey(password, salt)
	defer SecureZero(key) // Securely clear key from memory when done

	var nonce []byte
	var nonceSize int

	switch algorithm {
	case AlgoAES256_CTR:
		nonceSize = 16
	case AlgoChaCha20:
		nonceSize = 12
	default:
		return nil, fmt.Errorf("unsupported algorithm: %d", algorithm)
	}

	nonce = make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// 写入文件头
	header := new(bytes.Buffer)
	header.Write(magicHeader)
	header.WriteByte(currentVersion)
	header.WriteByte(byte(algorithm))
	header.WriteByte(byte(len(salt)))
	header.Write(salt)
	header.WriteByte(byte(len(nonce)))
	header.Write(nonce)

	// Version 2+: authenticate header to detect incorrect passwords early.
	if currentVersion >= version2 {
		header.Write(prf(key, header.Bytes()))
	}

	if _, err := w.Write(header.Bytes()); err != nil {
		return nil, err
	}

	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	return newParallelStreamWriter(w, algorithm, keyCopy, nonce)
}

// --- 解密读取器 ---

func NewDecryptedReader(r io.Reader, password string) (io.ReadCloser, error) {
	header := make([]byte, len(magicHeader))
	if _, err := io.ReadFull(r, header); err != nil {
		if err == io.EOF {
			return nil, ErrInvalidMagic // File too short
		}
		return nil, err
	}

	if !bytes.Equal(header, magicHeader) {
		return nil, ErrInvalidMagic
	}

	if password == "" {
		return nil, ErrPasswordRequired
	}

	// 读取版本和算法
	meta := make([]byte, 2)
	if _, err := io.ReadFull(r, meta); err != nil {
		return nil, err
	}
	versionByte := meta[0]
	algoByte := meta[1]
	if versionByte != version1 && versionByte != version2 {
		return nil, fmt.Errorf("unsupported encryption version: %d", versionByte)
	}

	// 读取 Salt
	saltLenBytes := make([]byte, 1)
	if _, err := io.ReadFull(r, saltLenBytes); err != nil {
		return nil, err
	}
	salt := make([]byte, saltLenBytes[0])
	if _, err := io.ReadFull(r, salt); err != nil {
		return nil, err
	}

	// 读取 Nonce/IV
	nonceLenBytes := make([]byte, 1)
	if _, err := io.ReadFull(r, nonceLenBytes); err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceLenBytes[0])
	if _, err := io.ReadFull(r, nonce); err != nil {
		return nil, err
	}

	key := deriveKey(password, salt)
	defer SecureZero(key)

	if versionByte >= version2 {
		expectedMac := make([]byte, sha256Size)
		if _, err := io.ReadFull(r, expectedMac); err != nil {
			return nil, err
		}

		hdr := new(bytes.Buffer)
		hdr.Write(magicHeader)
		hdr.WriteByte(versionByte)
		hdr.WriteByte(algoByte)
		hdr.WriteByte(saltLenBytes[0])
		hdr.Write(salt)
		hdr.WriteByte(nonceLenBytes[0])
		hdr.Write(nonce)

		actualMac := prf(key, hdr.Bytes())
		if !bytes.Equal(expectedMac, actualMac) {
			return nil, ErrInvalidPassword
		}
	}

	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	return newParallelStreamReaderWithPipe(r, algoByte, keyCopy, nonce)
}

// --- 实用工具函数 ---
// TODO
// 检查文件是否为加密文件
func IsEncryptedFile(r io.Reader) (bool, uint8, error) {
	header := make([]byte, len(magicHeader))
	n, err := r.Read(header)
	if err != nil || n < len(magicHeader) {
		return false, 0, nil
	}

	if !bytes.Equal(header, magicHeader) {
		return false, 0, nil
	}

	meta := make([]byte, 2)
	if _, err := r.Read(meta); err != nil {
		return false, 0, err
	}

	return true, meta[1], nil // 返回算法类型
}

// 安全清零内存
func SecureZero(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// 比较两个字节切片是否相等（防止时序攻击）
func ConstantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// 生成加密强度的随机字节
func GenerateSecureRandom(size int) ([]byte, error) {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return nil, fmt.Errorf("failed to generate random data: %w", err)
	}
	return data, nil
}

// 密码强度检查
func CheckPasswordStrength(password string) (score int, suggestions []string) {
	score = 0
	suggestions = []string{}

	length := len(password)
	if length < 8 {
		suggestions = append(suggestions, "使用至少8个字符，推荐12个以上")
	} else if length >= 10 {
		score += 2
	} else {
		score += 1
	}

	hasLower, hasUpper, hasDigit, hasSpecial := false, false, false, false
	for _, r := range password {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	complexity := 0
	if hasLower {
		complexity++
	} else {
		suggestions = append(suggestions, "包含小写字母")
	}
	if hasUpper {
		complexity++
	} else {
		suggestions = append(suggestions, "包含大写字母")
	}
	if hasDigit {
		complexity++
	} else {
		suggestions = append(suggestions, "包含数字")
	}
	if hasSpecial {
		complexity++
	} else {
		suggestions = append(suggestions, "包含特殊字符")
	}

	// 过短的密码不计入复杂度分，避免短密码因“类别齐全”而获得过高分数。
	if length >= 8 {
		score += complexity
	}

	// 检查常见弱密码模式
	weak_patterns := []string{"123456", "password", "qwerty", "admin", "root"}
	for _, pattern := range weak_patterns {
		if password == pattern {
			score -= 1
			suggestions = append(suggestions, "避免使用常见的弱密码模式")
			break
		}
	}

	if score < 0 {
		score = 0
	}
	if score > 6 {
		score = 6
	}

	return score, suggestions
}

func rotateRight32(x uint32, n int) uint32 {
	return bits.RotateLeft32(x, -n)
}
