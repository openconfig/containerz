package docker

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/docker/docker/client"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/volume"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
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

func (fakeDocker) ContainerInspect(ctx context.Context, container string) (types.ContainerJSON, error) {
	return types.ContainerJSON{}, fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerStart(ctx context.Context, container string, options container.StartOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) ContainerStop(ctx context.Context, container string, _ container.StopOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ImageLoad(ctx context.Context, input io.Reader, options ...client.ImageLoadOption) (image.LoadResponse, error) {
	return image.LoadResponse{}, fmt.Errorf("not implemented")
}

func (fakeDocker) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ImageRemove(ctx context.Context, image string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fakeDocker) ImageTag(ctx context.Context, source, target string) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) RegistryLogin(ctx context.Context, auth registry.AuthConfig) (registry.AuthenticateOKBody, error) {
	return registry.AuthenticateOKBody{}, fmt.Errorf("not implemented")
}

func (fakeDocker) VolumeCreate(ctx context.Context, options volume.CreateOptions) (volume.Volume, error) {
	return volume.Volume{}, fmt.Errorf("not implemented")
}

func (fakeDocker) VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error) {
	return volume.ListResponse{}, fmt.Errorf("not implemented")
}

func (fakeDocker) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) ContainersPrune(_ context.Context, _ filters.Args) (container.PruneReport, error) {
	return container.PruneReport{}, fmt.Errorf("not implemented")
}

func (fakeDocker) ImagesPrune(_ context.Context, _ filters.Args) (image.PruneReport, error) {
	return image.PruneReport{}, fmt.Errorf("not implemented")
}

func (fakeDocker) PluginCreate(ctx context.Context, createContext io.Reader, createOptions types.PluginCreateOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) PluginEnable(ctx context.Context, name string, options types.PluginEnableOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) PluginDisable(ctx context.Context, name string, options types.PluginDisableOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) PluginRemove(ctx context.Context, name string, options types.PluginRemoveOptions) error {
	return fmt.Errorf("not implemented")
}

func (fakeDocker) PluginList(ctx context.Context, filter filters.Args) (types.PluginsListResponse, error) {
	return types.PluginsListResponse{}, fmt.Errorf("not implemented")
}

func TestNew(t *testing.T) {
	want := &Manager{
		client: &fakeDocker{},
	}

	got := New(&fakeDocker{})

	opts := []cmp.Option{
		cmp.AllowUnexported(Manager{}),
		cmpopts.IgnoreFields(Manager{}, "janitor", "mu"),
		cmpopts.EquateEmpty(),
	}
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("New() returned diff (-want +got):\n%s", diff)
	}
}

func TestStop(t *testing.T) {
	d := &fakeDocker{}
	mgr := &Manager{
		client:  d,
		janitor: NewJanitor(d),
	}

	if err := mgr.Stop(context.Background()); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	if !d.CloseCalled {
		t.Errorf("Stop() did not close the underlying docker session.")
	}
}
