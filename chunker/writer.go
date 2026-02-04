// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package chunker is a convenience package to write and read files in chunks.
package chunker

import (
	"fmt"
	"os"
)

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
	written, err := w.tmp.WriteAt(p, int64(w.bytesWritten))
	if err != nil {
		return 0, err
	}

	w.chunkIndex++
	w.bytesWritten += uint64(written)
	return written, nil
}

func (w *Writer) Cleanup() error {
	w.tmp.Close()
	if err := os.Remove(w.tmp.Name()); err != nil {
		return fmt.Errorf("failed to remove temporary file %s with error: %s",
			w.tmp.Name(), err)
	}
	return nil
}

// Size returns the number of bytes written so far.
func (w Writer) Size() uint64 {
	return w.bytesWritten
}

// File returns the backing file where the data has been written.
func (w Writer) File() *os.File {
	return w.tmp
}
