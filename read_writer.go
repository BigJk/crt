package crt

import "io"

// ConcurrentRW is a concurrent read/write buffer via channels.
type ConcurrentRW struct {
	input  chan []byte
	output chan []byte
}

// NewConcurrentRW creates a new concurrent read/write buffer.
func NewConcurrentRW() *ConcurrentRW {
	return &ConcurrentRW{
		input:  make(chan []byte, 10),
		output: make(chan []byte),
	}
}

// Write writes data to the buffer.
func (rw *ConcurrentRW) Write(p []byte) (n int, err error) {
	data := make([]byte, len(p))
	copy(data, p)
	rw.input <- data
	return len(data), nil
}

// Read reads data from the buffer.
func (rw *ConcurrentRW) Read(p []byte) (n int, err error) {
	data, ok := <-rw.output
	if !ok {
		return 0, io.EOF
	}
	n = copy(p, data)
	return n, nil
}

// Run starts the concurrent read/write buffer.
func (rw *ConcurrentRW) Run() {
	const bufferSize = 1024
	buf := make([]byte, 0, bufferSize)
	for {
		select {
		case data, ok := <-rw.input:
			if !ok {
				close(rw.output)
				return
			}
			buf = append(buf, data...)
			for len(buf) > 0 {
				n := len(buf)
				if n > bufferSize {
					n = bufferSize
				}
				p := make([]byte, n)
				copy(p, buf[:n])
				buf = buf[n:]
				rw.output <- p
			}
		}
	}
}
