package engine

import (
	"bytes"
	"io"
	"sync"
)

type BodyCapture struct {
	body      io.ReadCloser
	limit     int
	mu        sync.Mutex
	buf       bytes.Buffer
	truncated bool
}

func NewBodyCapture(body io.ReadCloser, limit int) *BodyCapture {
	return &BodyCapture{body: body, limit: limit}
}

func (c *BodyCapture) Read(p []byte) (int, error) {
	n, err := c.body.Read(p)
	if n > 0 {
		c.mu.Lock()
		room := c.limit - c.buf.Len()
		switch {
		case room <= 0:
			c.truncated = true
		case n > room:
			c.buf.Write(p[:room])
			c.truncated = true
		default:
			c.buf.Write(p[:n])
		}
		c.mu.Unlock()
	}

	return n, err
}

func (c *BodyCapture) Close() error {
	return c.body.Close()
}

func (c *BodyCapture) Captured() ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return bytes.Clone(c.buf.Bytes()), c.truncated
}
