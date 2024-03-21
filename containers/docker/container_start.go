package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/v/v24/api/types/container"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/mount/mount"
	"github.com/moby/moby/v/v24/api/types/network"
	"github.com/moby/moby/v/v24/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
)

// ContainerStart starts a container provided the image exists and that the ports requested are not
// currently in use.
// TODO(alshabib) consider adding restart policy to containerz proto
func (m *Manager) ContainerStart(ctx context.Context, image, tag, cmd string, opts ...options.Option) (string, error) {
	optionz := options.ApplyOptions(opts...)

	images, err := m.client.ImageList(ctx, types.ImageListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return "", err
	}

	ref := fmt.Sprintf("%s:%s", image, tag)
	if err := findImage(ref, images); err != nil {
		return "", err
	}

	cnts, err := m.client.ContainerList(ctx, types.ContainerListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return "", err
	}

	if err := checkExistingInstanceAndPorts(optionz.InstanceName, optionz.PortMapping, cnts); err != nil {
		return "", err
	}

	mounts := make([]mount.Mount, 0, len(optionz.Volumes))
	for _, vol := range optionz.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:     "volume",
			Source:   vol.GetName(),
			Target:   vol.GetMountPoint(),
			ReadOnly: vol.GetReadOnly(),
		})
	}

	// TODO(alshabib): add resource support (i.e. CPU and memory quotas.)
	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		NetworkMode: "host", // TODO(alshabib): make this configurable.
		AutoRemove:  true,
	}
	config := &container.Config{
		Cmd:          strings.Split(cmd, " "),
		Image:        ref,
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,
		StdinOnce:    false,
		Tty:          true,
	}
	if len(optionz.PortMapping) > 0 {
		portMap := nat.PortMap{}
		portSet := nat.PortSet{}
		for in, out := range optionz.PortMapping {
			internal := fmt.Sprintf("%d", in)
			external := fmt.Sprintf("%d", out)
			in, err := nat.NewPort("tcp", internal)
			if err != nil {
				return "", err
			}

			portSet[in] = struct{}{}
			bindingV4 := nat.PortBinding{
				HostIP:   "0.0.0.0", // TODO(alshabib): do we want this to be configurable?
				HostPort: external,
			}
			bindingV6 := nat.PortBinding{
				HostIP:   "::",
				HostPort: external,
			}

			portMap[in] = []nat.PortBinding{bindingV4, bindingV6}
		}

		hostConfig.PortBindings = portMap
		config.ExposedPorts = portSet
	}

	if len(optionz.EnvMapping) > 0 {
		for envName, envVal := range optionz.EnvMapping {
			config.Env = append(config.Env, fmt.Sprintf("%s=%s", envName, envVal))
		}
	}

	resp, err := m.client.ContainerCreate(ctx, config, hostConfig, &network.NetworkingConfig{}, nil, optionz.InstanceName)
	if err != nil {
		return "", status.Errorf(codes.Internal, "unable to create container: %v", err)
	}

	if err := m.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", status.Errorf(codes.Internal, "unable to start container: %v", err)
	}

	name := resp.ID
	if optionz.InstanceName != "" {
		name = optionz.InstanceName
	}

	return name, nil
}

func checkExistingInstanceAndPorts(instance string, ports map[uint32]uint32, cnts []types.Container) error {
	if instance == "" && len(ports) == 0 {
		return nil
	}

	for _, cnt := range cnts {
		for _, name := range cnt.Names {
			strippedname := strings.Replace(name, "/", "", 1)
			if strippedname == instance {
				return status.Errorf(codes.AlreadyExists, "instance name %s already in use", instance)
			}
		}
		for _, port := range cnt.Ports {
			for _, ext := range ports {
				if ext == uint32(port.PublicPort) {
					return status.Errorf(codes.Unavailable, "port %d already in use", ext)
				}
			}
		}
	}
	return nil
}
