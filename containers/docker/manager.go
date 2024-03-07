// Package docker implements a container manager for docker orchestration.
package docker

import (
	"context"
	"io"

	"github.com/moby/moby/v/v24/api/types/container"
	"github.com/moby/moby/v/v24/api/types/filters"
	"github.com/moby/moby/v/v24/api/types/network"
	"github.com/moby/moby/v/v24/api/types/registry"
	"github.com/moby/moby/v/v24/api/types"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/volume/volume"

	ocispec "github.com/opencontainers/image-spec/tree/main/specs-go/v1"
)

type docker interface {
	Close() error
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, container string, options types.ContainerRemoveOptions) error
	ContainerStart(ctx context.Context, container string, options types.ContainerStartOptions) error
	ContainerStop(ctx context.Context, container string, options container.StopOptions) error
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	ImageLoad(ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error)
	ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error)
	ImageRemove(ctx context.Context, image string, options types.ImageRemoveOptions) ([]types.ImageDeleteResponseItem, error)
	ImageTag(ctx context.Context, source, target string) error
	RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error)
	VolumeCreate(ctx context.Context, options volume.CreateOptions) (volume.Volume, error)
	VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error

	ContainersPrune(ctx context.Context, args filters.Args) (types.ContainersPruneReport, error)
	ImagesPrune(ctx context.Context, args filters.Args) (types.ImagesPruneReport, error)
}

// Manager is a docker container orchestration manager.
type Manager struct {
	client  docker
	janitor *Vacuum
}

// New builds a new docker manager given a docker client.
func New(cli docker) *Manager {
	return &Manager{
		client:  cli,
		janitor: NewJanitor(cli),
	}
}

// Start starts a docker session to the host
func (m *Manager) Start(ctx context.Context) error {
	m.janitor.Start(ctx)
	return nil
}

// Stop closes the connection to the docker server.
func (m *Manager) Stop(ctx context.Context) error {
	m.janitor.Stop(ctx)
	return m.client.Close()
}
