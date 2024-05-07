package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
)

type fakeVacuumingDocker struct {
	fakeDocker
	cntCalled bool
	imgCalled bool
}

func (f *fakeVacuumingDocker) ContainersPrune(_ context.Context, _ filters.Args) (types.ContainersPruneReport, error) {
	f.cntCalled = true
	return types.ContainersPruneReport{}, nil
}

func (f *fakeVacuumingDocker) ImagesPrune(_ context.Context, _ filters.Args) (types.ImagesPruneReport, error) {
	f.imgCalled = true
	return types.ImagesPruneReport{}, nil
}

func TestVacuum(t *testing.T) {
	cleaningInterval = time.Second
	ctx := context.Background()

	fvd := &fakeVacuumingDocker{}
	jani := NewJanitor(fvd)
	jani.Start(ctx)

	time.Sleep(time.Second * 5)

	jani.Stop(ctx)

	if !fvd.cntCalled || !fvd.imgCalled {
		t.Errorf("Vacuum did not call the correct api: containers: %t, images: %t", fvd.cntCalled, fvd.imgCalled)
	}
}
