// Package docker implements a container manager for docker orchestration.
package docker

import (
	"context"
	"io"

	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/container/container"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/network/network"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/registry/registry"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/types"

	ocispec "google3/third_party/golang/opencontainers/image_spec/specs_go/v1/v1"
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
}

// Manager is a docker container orchestration manager.
type Manager struct {
	client docker
}

// New builds a new docker manager given a docker client.
func New(cli docker) *Manager {
	return &Manager{
		client: cli,
	}
}

// Stop closes the connection to the docker server.
func (m *Manager) Stop() error {
	return m.client.Close()
}
