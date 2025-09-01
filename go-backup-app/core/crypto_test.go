// core/crypto_test.go
package core

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptionDecryptionFlow(t *testing.T) {
	// 定义一个足够长的明文，以确保它跨越多个加密块
	longPlaintext := []byte(strings.Repeat("This is a secret message that is long enough to test multiple blocks. ", 10))
	password := "super_secret_password_123"

	testCases := []struct {
		name      string
		algo      uint8
		plaintext []byte
	}{
		{"AES256-CTR_Short", AlgoAES256_CTR, []byte("hello world")},
		{"AES256-CTR_Long", AlgoAES256_CTR, longPlaintext},
		{"ChaCha20_Short", AlgoChaCha20, []byte("hello world")},
		{"ChaCha20_Long", AlgoChaCha20, longPlaintext},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- Encryption ---
			ciphertextBuf := new(bytes.Buffer)

			// 1. 创建加密写入器
			encWriter, err := NewEncryptedWriter(ciphertextBuf, password, tc.algo)
			require.NoError(t, err)

			// 2. 写入明文
			n, err := encWriter.Write(tc.plaintext)
			require.NoError(t, err)
			assert.Equal(t, len(tc.plaintext), n)

			// 3. 关闭写入器
			err = encWriter.Close()
			require.NoError(t, err)

			// --- Decryption ---
			ciphertextReader := bytes.NewReader(ciphertextBuf.Bytes())

			// 1. 创建解密读取器
			decReader, err := NewDecryptedReader(ciphertextReader, password)
			require.NoError(t, err)

			// 2. 读取解密后的明文
			decryptedText, err := io.ReadAll(decReader)
			require.NoError(t, err)

			// 3. 验证结果
			assert.Equal(t, tc.plaintext, decryptedText, "Decrypted text should match original plaintext")

			// --- Negative Test: Wrong Password ---
			t.Run("WrongPassword", func(t *testing.T) {
				wrongPassReader := bytes.NewReader(ciphertextBuf.Bytes())
				decReaderWrong, err := NewDecryptedReader(wrongPassReader, "wrong_password")
				require.NoError(t, err) // Reader creation should succeed

				// Reading should produce garbled data, not the plaintext
				garbledText, err := io.ReadAll(decReaderWrong)
				require.NoError(t, err)
				assert.NotEqual(t, tc.plaintext, garbledText, "Text decrypted with wrong password should not match original")
			})
		})
	}
}

func TestHybridEncryptionDecryption(t *testing.T) {
	plaintext := []byte("This is a test of post-quantum hybrid encryption!")

	// --- Encryption ---
	ciphertextBuf := new(bytes.Buffer)

	// 1. 使用专为混合模式设计的辅助函数进行加密
	// 它会返回用于解密的私钥
	encWriter, privateKey, err := NewEncryptedWriterForHybrid(ciphertextBuf, AlgoHybrid)
	require.NoError(t, err)
	require.NotNil(t, privateKey)

	// 2. 写入明文
	_, err = encWriter.Write(plaintext)
	require.NoError(t, err)
	err = encWriter.Close()
	require.NoError(t, err)

	// --- Decryption ---
	ciphertextReader := bytes.NewReader(ciphertextBuf.Bytes())

	// 1. 使用对应的解密辅助函数和私钥进行解密
	decReader, err := NewDecryptedReaderForHybrid(ciphertextReader, privateKey)
	require.NoError(t, err)

	// 2. 读取解密后的明文
	decryptedText, err := io.ReadAll(decReader)
	require.NoError(t, err)

	// 3. 验证结果
	assert.Equal(t, plaintext, decryptedText)
}

func TestDilithiumSigning(t *testing.T) {
	message := []byte("This message will be signed by Dilithium mode 3.")

	// 1. 生成密钥对
	signer, err := GenerateDilithiumKeyPair()
	require.NoError(t, err)

	// 2. 使用私钥签名
	signature, err := signer.Sign(message)
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	// 3. 使用公钥验证签名 (成功场景)
	isValid := signer.Verify(message, signature)
	assert.True(t, isValid, "Signature should be valid with the correct message and key")

	// 4. 验证失败场景
	// 4a. 消息被篡改
	tamperedMessage := []byte("This message has been TAMPERED.")
	isTamperedValid := signer.Verify(tamperedMessage, signature)
	assert.False(t, isTamperedValid, "Signature should be invalid for a tampered message")

	// 4b. 使用错误的公钥
	otherSigner, err := GenerateDilithiumKeyPair()
	require.NoError(t, err)
	isOtherKeyValid := otherSigner.Verify(message, signature)
	assert.False(t, isOtherKeyValid, "Signature should be invalid with a different public key")
}

func TestInvalidMagicHeader(t *testing.T) {
	invalidData := []byte("NOT_A_VALID_FILE")
	r := bytes.NewReader(invalidData)

	_, err := NewDecryptedReader(r, "any_password")
	require.Error(t, err)
	assert.Equal(t, ErrInvalidMagic, err)
}
