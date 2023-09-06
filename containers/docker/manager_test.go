package docker

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/moby/moby/v/v24/api/types/container"
	"github.com/moby/moby/v/v24/api/types/network"
	"github.com/moby/moby/v/v24/api/types/registry"
	"github.com/moby/moby/v/v24/api/types"

	ocispec "github.com/opencontainers/image-spec/tree/main/specs-go/v1"
)

type fakeDocker struct {
	CloseCalled bool
}

func (f *fakeDocker) Close() error {
	f.CloseCalled = true
	return nil
}

func (fakeDocker) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	return container.CreateResponse{}, fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerRemove(ctx context.Context, container string, options types.ContainerRemoveOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerStart(ctx context.Context, container string, options types.ContainerStartOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerStop(ctx context.Context, container string, _ container.StopOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error) {
	return types.ImageLoadResponse{}, fmt.Errorf("not implemented")
}

func (fakeDocker) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ImageRemove(ctx context.Context, image string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ImageTag(ctx context.Context, source, target string) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
	return registry.AuthenticateOKBody{}, fmt.Errorf("not implemented")
}

func TestNew(t *testing.T) {
	want := &Manager{
		client: &fakeDocker{},
	}

	got := New(&fakeDocker{})

	if diff := cmp.Diff(want, got, cmp.AllowUnexported(Manager{})); diff != "" {
		t.Errorf("New() returned diff (-want +got):\n%s", diff)
	}
}

func TestStop(t *testing.T) {
	d := &fakeDocker{}
	mgr := &Manager{
		client: d,
	}

	if err := mgr.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	if !d.CloseCalled {
		t.Errorf("Stop() did not close the underlying docker session.")
	}
}
