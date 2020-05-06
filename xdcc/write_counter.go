package xdcc

import (
	"sync/atomic"
)

// WriteCounter counts the number of bytes written to it.
type WriteCounter struct {
	Total uint64
}

// Write implements the io.Writer interface.
//
// Always completes and never returns an error.
func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	atomic.AddUint64(&wc.Total, uint64(n))
	return n, nil
}
