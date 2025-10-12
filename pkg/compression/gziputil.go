package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// Compress сжимает данные с помощью gzip.
func Compress(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    zw := gzip.NewWriter(&buf)
    if _, err := zw.Write(data); err != nil {
        return nil, fmt.Errorf("ошибка записи gzip: %w", err)
    }
    if err := zw.Close(); err != nil {
        return nil, fmt.Errorf("ошибка закрытия gzip: %w", err)
    }
    return buf.Bytes(), nil
}

// Decompress распаковывает gzip-данные.
func Decompress(data []byte) ([]byte, error) {
    zr, err := gzip.NewReader(bytes.NewReader(data))
    if err != nil {
        return nil, fmt.Errorf("ошибка открытия gzip Reader: %w", err)
    }
    defer zr.Close()

    decompressed, err := io.ReadAll(zr)
    if err != nil {
        return nil, fmt.Errorf("ошибка чтения gzip данных: %w", err)
    }
    return decompressed, nil
}