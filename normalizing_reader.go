package scout

import (
	"errors"
	"io"
)

var _ io.Reader = (*NormalizingReader)(nil)

type NormalizingReader struct {
	io.Reader
}

// NewNormalizingReader creates a new NormalizingReader that captures
// all data read from the underlying reader.
func NewNormalizingReader(reader io.Reader) *NormalizingReader {
	return &NormalizingReader{Reader: reader}
}

// Read implements io.Reader by reading from the underlying reader
// and making sure there is a trailing newline.
// FUTURE: We may want to scan one line at a time and validate each line is valid JSONRPC, but that will have performance penalties
func (c *NormalizingReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		goto end
	}
	if n == len(p) {
		p = append(p, 0)
	}
	if p[n] != '\n' {
		p[n] = '\n'
	}
end:
	return n, err
}
