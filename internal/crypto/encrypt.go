package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io"
)

// Encrypt шифрует данные гибридной схемой: AES-GCM + RSA-OAEP.
// Формат результата: [2 байта длины RSA-блока][RSA-блок][nonce][зашифрованные данные]
func Encrypt(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	// 1. Генерируем случайный AES-256 ключ
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("generate aes key: %w", err)
	}

	// 2. Шифруем AES-ключ публичным RSA ключом
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("encrypt aes key: %w", err)
	}

	// 3. Шифруем данные через AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	encryptedData := gcm.Seal(nonce, nonce, data, nil)

	// 4. Собираем результат: [2 байта длины RSA][RSA-блок][nonce+данные]
	keyLen := len(encryptedAESKey)
	result := make([]byte, 2+keyLen+len(encryptedData))
	result[0] = byte(keyLen >> 8)
	result[1] = byte(keyLen)
	copy(result[2:], encryptedAESKey)
	copy(result[2+keyLen:], encryptedData)

	return result, nil
}
