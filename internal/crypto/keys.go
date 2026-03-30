package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadPublicKey — для агента
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	if path == "" {
		return nil, nil
	}

	publicKeyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load public key: %w", err)
	}

	publicKeyPemBlock, _ := pem.Decode(publicKeyBytes)
	if publicKeyPemBlock == nil {
		return nil, fmt.Errorf("load public key: no PEM data")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(publicKeyPemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("load public key: %w", err)
	}

	publicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("load public key: not RSA key")
	}

	return publicKey, nil

}

// LoadPrivateKey — для сервера
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	if path == "" {
		return nil, nil
	}

	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("load private key: no PEM data")
	}

	// Пробуем PKCS#8 (современный формат, генерируется по умолчанию)
	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaKey, ok := keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("load private key: not RSA key")
		}
		return rsaKey, nil
	}

	// Fallback на PKCS#1 (старый формат)
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}

	return rsaKey, nil
}
