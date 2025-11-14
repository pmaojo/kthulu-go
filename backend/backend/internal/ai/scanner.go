package ai

import (
	"bufio"
	"bytes"
	"io"
)

// BufferedLineScanner wraps bufio.Scanner for reading lines from streaming responses
type BufferedLineScanner struct {
	scanner *bufio.Scanner
	err     error
}

// NewBufferedLineScanner creates a new buffered line scanner
func NewBufferedLineScanner(r io.Reader) *BufferedLineScanner {
	return &BufferedLineScanner{
		scanner: bufio.NewScanner(r),
	}
}

// Scan reads the next line
func (s *BufferedLineScanner) Scan() bool {
	return s.scanner.Scan()
}

// Bytes returns the current line as bytes
func (s *BufferedLineScanner) Bytes() []byte {
	return bytes.TrimSpace(s.scanner.Bytes())
}

// Text returns the current line as string
func (s *BufferedLineScanner) Text() string {
	return s.scanner.Text()
}

// Err returns any error that occurred
func (s *BufferedLineScanner) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.scanner.Err()
}
