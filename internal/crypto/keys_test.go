package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func writePEMFile(t *testing.T, pemType string, der []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.pem")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: pemType, Bytes: der}); err != nil {
		t.Fatalf("pem encode: %v", err)
	}
	return f.Name()
}

func generateKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	return priv, &priv.PublicKey
}

// ── LoadPublicKey ────────────────────────────────────────────────────────────

func TestLoadPublicKey_EmptyPath(t *testing.T) {
	k, err := LoadPublicKey("")
	if err != nil || k != nil {
		t.Errorf("want (nil, nil), got (%v, %v)", k, err)
	}
}

func TestLoadPublicKey_ValidKey(t *testing.T) {
	_, pub := generateKeyPair(t)

	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	path := writePEMFile(t, "PUBLIC KEY", der)

	loaded, err := LoadPublicKey(path)
	if err != nil {
		t.Fatalf("LoadPublicKey: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil key")
	}
	if loaded.N.Cmp(pub.N) != 0 {
		t.Error("loaded key does not match original")
	}
}

func TestLoadPublicKey_FileNotFound(t *testing.T) {
	_, err := LoadPublicKey(filepath.Join(t.TempDir(), "nonexistent.pem"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadPublicKey_InvalidPEM(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.pem")
	f.WriteString("not a pem file")
	f.Close()

	_, err := LoadPublicKey(f.Name())
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

// ── LoadPrivateKey ───────────────────────────────────────────────────────────

func TestLoadPrivateKey_EmptyPath(t *testing.T) {
	k, err := LoadPrivateKey("")
	if err != nil || k != nil {
		t.Errorf("want (nil, nil), got (%v, %v)", k, err)
	}
}

func TestLoadPrivateKey_PKCS8(t *testing.T) {
	priv, _ := generateKeyPair(t)

	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	path := writePEMFile(t, "PRIVATE KEY", der)

	loaded, err := LoadPrivateKey(path)
	if err != nil {
		t.Fatalf("LoadPrivateKey PKCS8: %v", err)
	}
	if loaded.N.Cmp(priv.N) != 0 {
		t.Error("loaded key does not match original")
	}
}

func TestLoadPrivateKey_PKCS1(t *testing.T) {
	priv, _ := generateKeyPair(t)

	der := x509.MarshalPKCS1PrivateKey(priv)
	path := writePEMFile(t, "RSA PRIVATE KEY", der)

	loaded, err := LoadPrivateKey(path)
	if err != nil {
		t.Fatalf("LoadPrivateKey PKCS1: %v", err)
	}
	if loaded.N.Cmp(priv.N) != 0 {
		t.Error("loaded key does not match original")
	}
}

func TestLoadPrivateKey_FileNotFound(t *testing.T) {
	_, err := LoadPrivateKey(filepath.Join(t.TempDir(), "nonexistent.pem"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadPrivateKey_InvalidPEM(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.pem")
	f.WriteString("not pem")
	f.Close()

	_, err := LoadPrivateKey(f.Name())
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}
