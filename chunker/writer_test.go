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
	"math"
	"os"
	"testing"
)

func TestNewWriter(t *testing.T) {
	w, err := NewWriter(os.TempDir(), 10)
	if err != nil {
		t.Fatalf("NewWriter(%q, 10) returned an error: %v", os.TempDir(), err)
	}

	if w.tmp == nil || w.chunkSize != 10 || w.chunkIndex != 0 {
		t.Errorf("NewWriter(%q, 10) set incorrect fields: %+v", os.TempDir(), w)
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name        string
		inChunks    []byte
		inChunkSize int
		want        string
	}{
		{
			name:        "empty",
			inChunkSize: 1,
		},
		{
			name:        "small-chunks",
			inChunkSize: 1,
			inChunks:    []byte("containerz will contain all containers"),
			want:        "containerz will contain all containers",
		},
		{
			name:        "big-chunks",
			inChunkSize: 10,
			inChunks:    []byte("containerz will contain all containers in the world"),
			want:        "containerz will contain all containers in the world",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, err := NewWriter(os.TempDir(), tc.inChunkSize)
			if err != nil {
				t.Fatalf("NewWriter(%q, %v) returned an error: %v", os.TempDir(), tc.inChunkSize, err)
			}

			i := 0
			for {
				right := int(math.Min(float64(len(tc.inChunks)), float64((i+1)*tc.inChunkSize)))
				if _, err := w.Write(tc.inChunks[i*tc.inChunkSize : right]); err != nil {
					t.Errorf("Write(%q) returned an unexpected error: %v", tc.inChunks[i*tc.inChunkSize:right], err)
				}
				if right == len(tc.inChunks) {
					break
				}
				i++
			}

			gotBytes, err := io.ReadAll(w.File())
			if err != nil {
				t.Errorf("ReadAll(%q) returned an unexpected error: %v", w.File().Name(), err)
			}

			if string(gotBytes) != tc.want {
				t.Errorf("Write(%q) = %q, want: %q", string(tc.inChunks), string(gotBytes), tc.want)
			}

			if uint64(len(tc.want)) != w.Size() {
				t.Errorf("Write(%q) = %d, want: %d", string(tc.inChunks), w.Size(), len(tc.want))
			}

			if err := w.Cleanup(); err != nil {
				t.Errorf("Cleanup() returned an unexpected error: %v", err)
			}
			if _, err := os.Stat(w.File().Name()); !os.IsNotExist(err) {
				t.Errorf("Cleanup() did not remove the temporary file: %s", w.File().Name())
			}
		})
	}
}
