package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
)

type fakeVacuumingDocker struct {
	fakeDocker
	cntCalled bool
	imgCalled bool
}

func (f *fakeVacuumingDocker) ContainersPrune(_ context.Context, _ filters.Args) (container.PruneReport, error) {
	f.cntCalled = true
	return container.PruneReport{}, nil
}

func (f *fakeVacuumingDocker) ImagesPrune(_ context.Context, _ filters.Args) (image.PruneReport, error) {
	f.imgCalled = true
	return image.PruneReport{}, nil
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
