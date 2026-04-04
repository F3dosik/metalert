package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey

	original := []byte(`[{"id":"Alloc","type":"gauge","value":123.45}]`)

	encrypted, err := Encrypt(original, publicKey)
	require.NoError(t, err)

	decrypted, err := Decrypt(encrypted, privateKey)
	require.NoError(t, err)

	assert.Equal(t, original, decrypted)
}
