package chunker

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewReader(t *testing.T) {
	file := "testdata/reader-data.txt"
	r, err := NewReader(file)
	if err != nil {
		t.Fatalf("NewReader(%q) returned error: %v", file, err)
	}

	if r.f == nil || r.fileSize != 26 || r.chunkIndex != 0 {
		t.Errorf("NewReader(%q) set incorrect fields: %+v", file, r)
	}
}

func TestRead(t *testing.T) {
	tests := []struct {
		name      string
		chunkSize int32
		want      string
	}{
		{
			name:      "tiny-chunks",
			chunkSize: 1,
			want:      "some really important data",
		},
		{
			name:      "bug-chunks",
			chunkSize: 4,
			want:      "some really important data",
		},
	}

	file := "testdata/reader-data.txt"
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewReader(file)
			if err != nil {
				t.Fatalf("NewReader(%q) returned error: %v", file, err)
			}

			var res string
			for {
				got, err := r.Read(tc.chunkSize)
				if err != nil {
					if err == io.EOF {
						break
					}
					t.Errorf("Read(%v) returned an unexpected error: %v", tc.chunkSize, err)
				}

				res += string(got)
			}

			if diff := cmp.Diff(tc.want, res); diff != "" {
				t.Errorf("Read(%v) returned an unexpected diff (-want +got): %v", tc.chunkSize, diff)
			}
		})
	}
}
