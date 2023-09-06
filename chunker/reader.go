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

package chunker

import (
	"io"
	"os"
)

// Reader reads a file in chunks.
type Reader struct {
	f          *os.File
	fileSize   uint64
	chunkIndex int32
	done       bool
}

// NewReader builds a new chunked reader.
func NewReader(file string) (*Reader, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &Reader{
		f:        f,
		fileSize: uint64(stat.Size()),
	}, nil
}

// Read reads a chunk of the file.
func (r *Reader) Read(chunkSize int32) ([]byte, error) {
	if r.done {
		return nil, io.EOF
	}

	buf := make([]byte, chunkSize)

	n, err := r.f.ReadAt(buf, int64(r.chunkIndex*chunkSize))
	if err != nil {
		if err == io.EOF {
			r.done = true
			return buf[:n], nil
		}
		return nil, err
	}

	r.chunkIndex++
	return buf, nil
}

// Size returns the size of the file.
func (r Reader) Size() uint64 {
	return r.fileSize
}

// Close closes the file
func (r *Reader) Close() error {
	return r.f.Close()
}

// IsEOF indicates that there is nothing more in the file.
func (r Reader) IsEOF() bool {
	return r.done
}
