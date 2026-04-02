package gzip

import (
	"io"
	"net/http"
	"testing"
)

type discardResponseWriter struct{}

func (d *discardResponseWriter) Header() http.Header         { return http.Header{} }
func (d *discardResponseWriter) Write(p []byte) (int, error) { return io.Discard.Write(p) }
func (d *discardResponseWriter) WriteHeader(statusCode int)  {}

func BenchmarkCompressWriter_WithPool(b *testing.B) {
	body := []byte(`{"id":"Alloc","type":"gauge","value":123.45}`)
	w := &discardResponseWriter{}
	b.ReportAllocs()

	for b.Loop() {
		cw := newCompressWriter(w)
		cw.Write(body)
		cw.Close()
	}
}
