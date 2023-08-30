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
