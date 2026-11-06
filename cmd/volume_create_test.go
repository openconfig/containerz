// Copyright 2026 Google LLC
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

package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseKeyValues(t *testing.T) {
	tests := []struct {
		name    string
		in      []string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "empty",
			want: map[string]string{},
		},
		{
			name: "multiple-entries",
			in:   []string{"options=bind", "mountpoint=/data"},
			want: map[string]string{"options": "bind", "mountpoint": "/data"},
		},
		{
			name: "value-with-commas",
			in:   []string{"options=rbind,rslave,ro"},
			want: map[string]string{"options": "rbind,rslave,ro"},
		},
		{
			name: "value-with-equals",
			in:   []string{"options=rbind,uid=1000"},
			want: map[string]string{"options": "rbind,uid=1000"},
		},
		{
			name: "empty-value",
			in:   []string{"type="},
			want: map[string]string{"type": ""},
		},
		{
			name:    "missing-equals",
			in:      []string{"foo"},
			wantErr: true,
		},
		{
			name:    "empty-key",
			in:      []string{"=bar"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseKeyValues("option", tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("parseKeyValues(%q) returned error %v, wantErr %t", tc.in, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseKeyValues(%q) returned diff (-want, +got):\n%s", tc.in, diff)
			}
		})
	}
}
