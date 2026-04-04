package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

// Decrypt расшифровывает данные зашифрованные через Encrypt.
func Decrypt(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("decrypt: data too short")
	}

	// 1. Читаем длину RSA-блока
	keyLen := int(data[0])<<8 | int(data[1])
	if len(data) < 2+keyLen {
		return nil, fmt.Errorf("decrypt: invalid data format")
	}

	// 2. Расшифровываем AES-ключ
	encryptedAESKey := data[2 : 2+keyLen]
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedAESKey, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt aes key: %w", err)
	}

	// 3. Расшифровываем данные
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}

	encryptedData := data[2+keyLen:]
	if len(encryptedData) < gcm.NonceSize() {
		return nil, fmt.Errorf("decrypt: encrypted data too short")
	}

	nonce := encryptedData[:gcm.NonceSize()]
	ciphertext := encryptedData[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt data: %w", err)
	}

	return plaintext, nil
}
