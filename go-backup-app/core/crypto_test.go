// core/crypto_test.go
package core

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"
)

// 测试自定义的 SHA256 实现是否与标准向量匹配
func TestCustomSHA256(t *testing.T) {
	testCases := []struct {
		name   string
		input  []byte
		output string
	}{
		{
			"empty string",
			[]byte(""),
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			"abc",
			[]byte("abc"),
			"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		},
		{
			"The quick brown fox jumps over the lazy dog",
			[]byte("The quick brown fox jumps over the lazy dog"),
			"d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash := Sum256(tc.input)
			hexHash := hex.EncodeToString(hash[:])
			if hexHash != tc.output {
				t.Errorf("Sum256() got = %v, want %v", hexHash, tc.output)
			}
		})
	}
}

// 测试 prf (HMAC-SHA256) 函数
func TestPRF_HMAC_SHA256(t *testing.T) {
	// Test vector from RFC 4231
	key := []byte("Jefe")
	data := []byte("what do ya want for nothing?")
	expected := "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843"

	mac := prf(key, data)
	hexMac := hex.EncodeToString(mac)

	if hexMac != expected {
		t.Errorf("prf() got = %v, want %v", hexMac, expected)
	}
}

// 测试 pbkdf2 函数
func TestPBKDF2(t *testing.T) {
	// Test vector from RFC 6070
	password := []byte("password")
	salt := []byte("salt")
	iter := 4096
	keyLen := 32
	expected := "c5e478d59288c841aa530db6845c4c8d962893a001ce4e11a4963873aa98134a"

	key := pbkdf2(password, salt, iter, keyLen)
	hexKey := hex.EncodeToString(key)

	if hexKey != expected {
		t.Errorf("pbkdf2() got = %v, want %v", hexKey, expected)
	}
}

// 测试完整的加密和解密往返流程
func TestEncryptionDecryptionCycle(t *testing.T) {
	plaintext := []byte("This is a secret message that needs to be encrypted and then decrypted successfully.")
	password := "my-very-strong-p@ssw0rd!123"

	algorithms := []struct {
		name string
		id   uint8
	}{
		{"AES-256-CTR", AlgoAES256_CTR},
		{"ChaCha20", AlgoChaCha20},
	}

	for _, algo := range algorithms {
		t.Run(algo.name, func(t *testing.T) {
			var encryptedData bytes.Buffer

			// 加密
			writer, err := NewEncryptedWriter(&encryptedData, password, algo.id)
			if err != nil {
				t.Fatalf("NewEncryptedWriter() error = %v", err)
			}
			_, err = writer.Write(plaintext)
			if err != nil {
				t.Fatalf("writer.Write() error = %v", err)
			}
			err = writer.Close()
			if err != nil {
				t.Fatalf("writer.Close() error = %v", err)
			}

			// 解密
			reader, err := NewDecryptedReader(&encryptedData, password)
			if err != nil {
				t.Fatalf("NewDecryptedReader() error = %v", err)
			}

			decryptedText, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("io.ReadAll() error = %v", err)
			}

			// 验证
			if !bytes.Equal(plaintext, decryptedText) {
				t.Errorf("Decrypted text does not match original plaintext.\nOriginal: %s\nDecrypted: %s", plaintext, decryptedText)
			}
		})
	}
}
