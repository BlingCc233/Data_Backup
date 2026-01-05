package core

import (
	"bytes"
	"crypto/aes"
	standard_sha256 "crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"golang.org/x/crypto/chacha20"
)

// --- SHA-256 测试 ---

// TestSum256Vectors 使用已知的测试向量来验证 SHA256 实现
func TestSum256Vectors(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty String",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "abc",
			input:    "abc",
			expected: "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		},
		{
			name:     "long string",
			input:    "abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq",
			expected: "248d6a61d20638b8e5c026930c3e6039a33ce45964ff2167f6ecedd419db06c1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := Sum256([]byte(tc.input))
			got := hex.EncodeToString(hash[:])
			if got != tc.expected {
				t.Errorf("Sum256(%q) = %s; want %s", tc.input, got, tc.expected)
			}
		})
	}
}

// TestSHA256AgainstStdlib 将 SHA256 实现与标准库进行比较
func TestSHA256AgainstStdlib(t *testing.T) {
	// 使用随机数据进行测试
	data := make([]byte, 1024*1024) // 1MB
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(data)

	// 测试不同的数据切片大小
	sizes := []int{0, 1, 63, 64, 65, 1000, 1024 * 1024}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("data_size_%d", size), func(t *testing.T) {
			input := data[:size]

			// 实现
			customHash := Sum256(input)

			// 标准库实现
			stdHash := standard_sha256.Sum256(input)

			if !bytes.Equal(customHash[:], stdHash[:]) {
				t.Errorf("Hashes do not match for size %d", size)
				t.Errorf("Custom: %x", customHash)
				t.Errorf("Stdlib: %x", stdHash)
			}
		})
	}
}

// TestDigestWrite 将 digest writer 与标准库进行比较
func TestDigestWrite(t *testing.T) {
	customDigest := New()
	stdDigest := standard_sha256.New()

	data := []byte("this is a test of writing data in chunks to the digest")

	// 模拟分块写入
	chunkSize := 5
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]
		customDigest.Write(chunk)
		stdDigest.Write(chunk)
	}

	customSum := customDigest.Sum(nil)
	stdSum := stdDigest.Sum(nil)

	if !bytes.Equal(customSum, stdSum) {
		t.Errorf("Digest sums do not match after chunked write")
		t.Errorf("Custom: %x", customSum)
		t.Errorf("Stdlib: %x", stdSum)
	}
}

// --- PBKDF2 (密钥派生) 测试 ---

// TestPBKDF2Vectors 使用 PBKDF2-HMAC-SHA256 的测试向量（对齐 x/crypto/pbkdf2）
func TestPBKDF2Vectors(t *testing.T) {
	testCases := []struct {
		name        string
		password    string
		salt        string
		iter        int
		keyLen      int
		expectedKey string
	}{
		{
			name:        "PBKDF2_SHA256_password_salt_4096",
			password:    "password",
			salt:        "salt",
			iter:        4096,
			keyLen:      25,
			expectedKey: "c5e478d59288c841aa530db6845c4c8d962893a001ce4e11a4",
		},
		{
			// 自定义测试，匹配 deriveKey 函数
			name:        "Custom_deriveKey_test",
			password:    "my-secret-password-123",
			salt:        "some-random-salt",
			iter:        4096,
			keyLen:      32,
			expectedKey: "ab8ed9f94dd4defe768708082330c57b8a5b867e5bea107b5a8205a7526c7054",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := pbkdf2([]byte(tc.password), []byte(tc.salt), tc.iter, tc.keyLen)
			got := hex.EncodeToString(key)
			if got != tc.expectedKey {
				t.Errorf("pbkdf2() = %s; want %s", got, tc.expectedKey)
			}
		})
	}
}

// --- AES 测试 ---

// TestAESAgainstStdlib 将 AES 加密与标准库进行比较
func TestAESAgainstStdlib(t *testing.T) {
	keySizes := []int{16, 24, 32} // AES-128, AES-192, AES-256
	plaintext := []byte("this is 16 bytes")

	for _, keySize := range keySizes {
		t.Run(fmt.Sprintf("AES-%d", keySize*8), func(t *testing.T) {
			key := make([]byte, keySize)
			rand.Read(key)

			// 实现
			customAES, err := NewAES(key)
			if err != nil {
				t.Fatalf("NewAES failed: %v", err)
			}
			customCiphertext := customAES.Encrypt(plaintext)

			// 标准库实现
			stdAES, err := aes.NewCipher(key)
			if err != nil {
				t.Fatalf("aes.NewCipher failed: %v", err)
			}
			stdCiphertext := make([]byte, 16)
			stdAES.Encrypt(stdCiphertext, plaintext)

			if !bytes.Equal(customCiphertext, stdCiphertext) {
				t.Errorf("AES encryption results do not match")
				t.Errorf("Custom: %x", customCiphertext)
				t.Errorf("Stdlib: %x", stdCiphertext)
			}
		})
	}
}

// --- ChaCha20 测试 ---

// TestChaCha20AgainstCryptoLib 将 ChaCha20 块生成与 x/crypto 库进行比较
func TestChaCha20AgainstCryptoLib(t *testing.T) {
	key := make([]byte, 32)
	nonce := make([]byte, 12)
	rand.Read(key)
	rand.Read(nonce)

	// 实现
	customChaCha, err := NewChaCha20(key, nonce)
	if err != nil {
		t.Fatalf("NewChaCha20 failed: %v", err)
	}
	customChaCha.counter = 1
	customBlock := customChaCha.block()

	// x/crypto 实现
	stdChaCha, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		t.Fatalf("chacha20.NewUnauthenticatedCipher failed: %v", err)
	}
	stdChaCha.SetCounter(1)
	stdBlock := make([]byte, 64)
	zeros := make([]byte, 64)
	stdChaCha.XORKeyStream(stdBlock, zeros) // 加密64字节的0以获得密钥流

	if !bytes.Equal(customBlock[:], stdBlock) {
		t.Errorf("ChaCha20 block results do not match")
		t.Errorf("Custom: %x", customBlock)
		t.Errorf("Stdlib: %x", stdBlock)
	}
}

// --- 流密码和端到端测试 ---

func TestParallelEncryptionDecryption(t *testing.T) {
	runE2ETest := func(t *testing.T, algorithm uint8, dataSize int) {
		// 1. 准备数据和密码
		originalData := make([]byte, dataSize)
		_, err := rand.Read(originalData)
		if err != nil {
			t.Fatalf("Failed to generate random data: %v", err)
		}
		password := "a-very-strong-and-secure-password"

		// 2. 加密
		encryptedBuf := new(bytes.Buffer)
		writer, err := NewEncryptedWriter(encryptedBuf, password, algorithm)
		if err != nil {
			t.Fatalf("NewEncryptedWriter failed: %v", err)
		}

		n, err := io.Copy(writer, bytes.NewReader(originalData))
		if err != nil {
			t.Fatalf("io.Copy to writer failed: %v", err)
		}
		if n != int64(dataSize) {
			t.Fatalf("Expected to write %d bytes, but wrote %d", dataSize, n)
		}

		err = writer.Close()
		if err != nil {
			t.Fatalf("Writer.Close() failed: %v", err)
		}

		// 3. 解密
		reader, err := NewDecryptedReader(encryptedBuf, password)
		if err != nil {
			t.Fatalf("NewDecryptedReader failed: %v", err)
		}
		defer reader.Close()

		decryptedData, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("io.ReadAll from reader failed: %v", err)
		}

		// 4. 验证
		if len(decryptedData) != len(originalData) {
			t.Fatalf("Length mismatch: original %d, decrypted %d", len(originalData), len(decryptedData))
		}

		if !bytes.Equal(originalData, decryptedData) {
			t.Fatal("Decrypted data does not match original data")
		}
		t.Logf("Successfully encrypted and decrypted %d bytes", dataSize)
	}

	testCases := []struct {
		algoName  string
		algoCode  uint8
		dataSize  int
		dataLabel string
	}{
		{"AES_small", AlgoAES256_CTR, 512 * 1024, "512KB"},       // 小于一个 chunk
		{"AES_large", AlgoAES256_CTR, 3 * 1024 * 1024, "3MB"},    // 多个 chunk
		{"ChaCha20_small", AlgoChaCha20, 512 * 1024, "512KB"},    // 小于一个 chunk
		{"ChaCha20_large", AlgoChaCha20, 3 * 1024 * 1024, "3MB"}, // 多个 chunk
		{"ChaCha20_exact", AlgoChaCha20, 2 * 1024 * 1024, "2MB"}, // 正好是 chunk 的整数倍
		{"AES_tiny", AlgoAES256_CTR, 100, "100B"},                // 远小于一个 chunk
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.algoName, tc.dataLabel), func(t *testing.T) {
			runE2ETest(t, tc.algoCode, tc.dataSize)
		})
	}
}

// --- 工具函数测试 ---

func TestSecureZero(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	SecureZero(data)
	expected := make([]byte, 8)
	if !bytes.Equal(data, expected) {
		t.Errorf("SecureZero failed. Expected all zeros, got %v", data)
	}
}

func TestConstantTimeCompare(t *testing.T) {
	a := []byte("hello")
	b := []byte("hello")
	c := []byte("world")
	d := []byte("hellp")
	e := []byte("helloworld")

	if !ConstantTimeCompare(a, b) {
		t.Error("Expected a and b to be equal")
	}
	if ConstantTimeCompare(a, c) {
		t.Error("Expected a and c to be unequal")
	}
	if ConstantTimeCompare(a, d) {
		t.Error("Expected a and d to be unequal")
	}
	if ConstantTimeCompare(a, e) {
		t.Error("Expected a and e to be unequal (different length)")
	}
}

func TestCheckPasswordStrength(t *testing.T) {
	testCases := []struct {
		password  string
		wantScore int
		// 不严格检查建议，只检查分数
	}{
		{"123", 0},
		{"password", 1},
		{"Password", 3},
		{"Password123", 5},
		{"P@ssword123!", 6},
		{"thisisalongpassword", 3},
		{"ThisIsAVeryLongAndComplexP@ssword123!", 6},
	}

	for _, tc := range testCases {
		t.Run(tc.password, func(t *testing.T) {
			score, _ := CheckPasswordStrength(tc.password)
			if score != tc.wantScore {
				t.Errorf("CheckPasswordStrength(%q) score = %d; want %d", tc.password, score, tc.wantScore)
			}
		})
	}
}

// --- 基准测试 ---

var (
	benchData1MB = make([]byte, 1024*1024)
	benchKey32   = make([]byte, 32)
	benchNonce16 = make([]byte, 16)
)

func init() {
	rand.Read(benchData1MB)
	rand.Read(benchKey32)
	rand.Read(benchNonce16)
}

func BenchmarkSum256Custom(b *testing.B) {
	b.SetBytes(int64(len(benchData1MB)))
	for i := 0; i < b.N; i++ {
		Sum256(benchData1MB)
	}
}

func BenchmarkSum256Stdlib(b *testing.B) {
	b.SetBytes(int64(len(benchData1MB)))
	for i := 0; i < b.N; i++ {
		standard_sha256.Sum256(benchData1MB)
	}
}

func BenchmarkAESEncryptCustom(b *testing.B) {
	aes, err := NewAES(benchKey32)
	if err != nil {
		b.Fatal(err)
	}
	plaintext := make([]byte, 16)
	b.SetBytes(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		aes.Encrypt(plaintext)
	}
}

func BenchmarkAESEncryptStdlib(b *testing.B) {
	stdAES, err := aes.NewCipher(benchKey32)
	if err != nil {
		b.Fatal(err)
	}
	plaintext := make([]byte, 16)
	ciphertext := make([]byte, 16)
	b.SetBytes(16)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stdAES.Encrypt(ciphertext, plaintext)
	}
}

func BenchmarkParallelWriterAES(b *testing.B) {
	b.SetBytes(int64(len(benchData1MB)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer() // 暂停计时器以进行设置
		writer, err := newParallelStreamWriter(io.Discard, AlgoAES256_CTR, benchKey32, benchNonce16)
		if err != nil {
			b.Fatal(err)
		}
		reader := bytes.NewReader(benchData1MB)

		b.StartTimer() // 恢复计时器
		io.Copy(writer, reader)
		writer.Close()
	}
}

func BenchmarkParallelWriterChaCha20(b *testing.B) {
	nonce12 := make([]byte, 12)
	copy(nonce12, benchNonce16)

	b.SetBytes(int64(len(benchData1MB)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		writer, err := newParallelStreamWriter(io.Discard, AlgoChaCha20, benchKey32, nonce12)
		if err != nil {
			b.Fatal(err)
		}
		reader := bytes.NewReader(benchData1MB)

		b.StartTimer()
		io.Copy(writer, reader)
		writer.Close()
	}
}
