package compression

import (
	"bytes"
	"testing"
)

func TestCompressDecompress(t *testing.T) {
	original := []byte("hello, metrics world! " + string(make([]byte, 100)))

	compressed, err := Compress(original)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("compressed data is empty")
	}

	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	if !bytes.Equal(decompressed, original) {
		t.Errorf("roundtrip mismatch: got %q, want %q", decompressed, original)
	}
}

func TestCompressEmptyInput(t *testing.T) {
	compressed, err := Compress([]byte{})
	if err != nil {
		t.Fatalf("Compress empty: %v", err)
	}
	decompressed, err := Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress empty: %v", err)
	}
	if len(decompressed) != 0 {
		t.Errorf("expected empty result, got %d bytes", len(decompressed))
	}
}

func TestDecompressInvalidData(t *testing.T) {
	_, err := Decompress([]byte("not gzip data"))
	if err == nil {
		t.Error("expected error for invalid gzip data")
	}
}

func TestCompressReducesSize(t *testing.T) {
	// Повторяющиеся данные хорошо сжимаются.
	data := bytes.Repeat([]byte("aaaa"), 500)
	compressed, err := Compress(data)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if len(compressed) >= len(data) {
		t.Errorf("expected compressed size < original, got %d >= %d", len(compressed), len(data))
	}
}
