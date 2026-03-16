package gzip

import (
	"compress/gzip"
	"io"
	"sync"
)

var readerPool = sync.Pool{
	New: func() any {
		return &compressReader{}
	},
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	cr := readerPool.Get().(*compressReader)
	if cr.zr == nil {
		zr, err := gzip.NewReader(r)
		if err != nil {
			readerPool.Put(cr)
			return nil, err
		}
		cr.zr = zr
	} else {
		if err := cr.zr.Reset(r); err != nil {
			readerPool.Put(cr)
			return nil, err
		}
	}
	cr.r = r
	return cr, nil
}

func (c *compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	err1 := c.zr.Close()
	err2 := c.r.Close()
	c.r = nil
	readerPool.Put(c)
	if err1 != nil {
		return err1
	}
	return err2
}
