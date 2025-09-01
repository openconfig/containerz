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

package client

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNonBlockingSend(t *testing.T) {
	tests := []struct {
		name   string
		cancel bool
		ch     chan *Progress
		data   *Progress
		want   bool
	}{
		{
			name: "normal-send",
			ch:   make(chan *Progress),
			data: &Progress{},
			want: false,
		},
		{
			name:   "cancelled",
			ch:     make(chan *Progress),
			cancel: true,
			data:   &Progress{},
			want:   true,
		},
		{
			name:   "cancelled-with-ch-cap",
			ch:     make(chan *Progress, 10),
			cancel: true,
			data:   &Progress{},
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.cancel {
				cancel()
			}

			go func() {
				<-tc.ch
			}()

			got := nonBlockingChannelSend(ctx, tc.ch, tc.data)

			if got != tc.want {
				t.Errorf("nonBlockingSend returned wrong state want: %t, got: %t", tc.want, got)
			}
		})
	}
}

func TestWithEnv(t *testing.T) {
	o := &startOptions{}

	in := []string{"containerz"}
	WithEnv(in)(o)

	if diff := cmp.Diff(in, o.envs); diff != "" {
		t.Errorf("WithEnv(%v) returned diff (-want, +got):\n%s", in, diff)
	}
}

func TestWithPorts(t *testing.T) {
	o := &startOptions{}

	in := []string{"containerz"}
	WithPorts(in)(o)

	if diff := cmp.Diff(in, o.ports); diff != "" {
		t.Errorf("WithPorts(%v) returned diff (-want, +got):\n%s", in, diff)
	}
}
