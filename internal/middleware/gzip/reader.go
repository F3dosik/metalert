package gzip

import (
	"compress/gzip"
	"io"
)

type compressReader struct {
    r  io.ReadCloser
    zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
    zr, err := gzip.NewReader(r)
    if err != nil {
        return nil, err
    }
    return &compressReader{r: r, zr: zr}, nil
}

func (c *compressReader) Read(p []byte) (int, error) {
    return c.zr.Read(p)
}

func (c *compressReader) Close() error {
    err1 := c.zr.Close()
    err2 := c.r.Close()
    if err1 != nil {
        return err1
    }
    return err2
}
