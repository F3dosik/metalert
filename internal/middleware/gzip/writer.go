package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

var writerPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return &compressWriter{zw: w}
	},
}

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	cw := writerPool.Get().(*compressWriter)
	cw.zw.Reset(w)
	cw.w = w
	return cw
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	err := c.zw.Close()
	c.w = nil
	writerPool.Put(c) // возвращаем в пул
	return err
}

func (c *compressWriter) Flush() {
	c.zw.Flush()
	if f, ok := c.w.(http.Flusher); ok {
		f.Flush()
	}
}
