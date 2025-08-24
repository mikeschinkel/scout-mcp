package mcputil

import (
	"errors"
	"fmt"
	"io"
)

var _ io.Reader = (*CapturingReader)(nil)
var _ fmt.Stringer = (*CapturingReader)(nil)

// CapturingReader wraps an io.Reader to capture all read data
// for debugging and error reporting purposes.
type CapturingReader struct {
	io.Reader
	buffer    []byte
	bytesRead int64
}

// NewCapturingReader creates a new CapturingReader that captures
// all data read from the underlying reader.
func NewCapturingReader(reader io.Reader) *CapturingReader {
	return &CapturingReader{Reader: reader}
}

// Read implements io.Reader by reading from the underlying reader
// while capturing the read data for later inspection.
func (c *CapturingReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		goto end
	}
	c.bytesRead += int64(n)
	c.buffer = append(c.buffer, p...)
end:
	return n, err
}

func (c *CapturingReader) String() string {
	return string(c.Bytes())
}
func (c *CapturingReader) Bytes() []byte {
	return c.buffer[:c.bytesRead]
}
