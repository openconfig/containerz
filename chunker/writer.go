// Package chunker is a convenience package to write and read files in chunks.
package chunker

import "os"

// Writer is an implementation of a chunked writer.
type Writer struct {
	tmp        *os.File
	chunkSize  int
	chunkIndex int

	bytesWritten uint64
}

// NewWriter returns a chunked writer.
func NewWriter(location string, chunkSize int) (*Writer, error) {
	f, err := os.CreateTemp(location, "*")
	if err != nil {
		return nil, err
	}

	return &Writer{tmp: f, chunkSize: chunkSize, chunkIndex: 0}, nil
}

func (w *Writer) Write(p []byte) (int, error) {
	written, err := w.tmp.WriteAt(p, int64(w.chunkIndex*w.chunkSize))
	if err != nil {
		return 0, err
	}

	w.chunkIndex++
	w.bytesWritten += uint64(written)
	return written, nil
}

// Size returns the number of bytes written so far.
func (w Writer) Size() uint64 {
	return w.bytesWritten
}

// File returns the backing file where the data has been written.
func (w Writer) File() *os.File {
	return w.tmp
}
