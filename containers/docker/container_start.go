package docker

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"

	cpb "github.com/openconfig/gnoi/containerz"
)

// knownShells is a slice of shells which can be used in commands that containers
// can be started with. All of these shells can run commands using the -c flag, like:
// <shell> -c "<command>"
var knownShells = []string{
	"sh",
	"bash",
	"zsh",
	"ksh",
	"fish",
	"tcsh",
}

const commandFlag = "-c"

// ContainerStart starts a container provided the image exists and that the ports requested are not
// currently in use.
func (m *Manager) ContainerStart(ctx context.Context, image, tag, cmd string, opts ...options.Option) (string, error) {
	optionz := options.ApplyOptions(opts...)

	images, err := m.client.ImageList(ctx, imagetypes.ListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return "", err
	}

	ref := fmt.Sprintf("%s:%s", image, tag)
	if err := findImage(ref, images); err != nil {
		return "", err
	}

	cnts, err := m.client.ContainerList(ctx, container.ListOptions{
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

	cpu, err := options.ParseCPUs(optionz.CPU)
	if err != nil {
		return "", fmt.Errorf("unable to parse cpu limit %f: %v", optionz.CPU, err)
	}

	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		NetworkMode: "host",

		Resources: container.Resources{
			NanoCPUs:          cpu,
			Memory:            optionz.HardMemory, // hard
			MemoryReservation: optionz.SoftMemory, // soft
		},
	}
	splitCmd := strings.Split(cmd, " ")
	if len(splitCmd) > 2 &&
		splitCmd[1] == commandFlag &&
		slices.Contains(knownShells, splitCmd[0]) {
		// command is of the form <shell> -c "<command>"
		quoted := strings.TrimPrefix(cmd, splitCmd[0]+" "+commandFlag+" ")
		unquoted, err := strconv.Unquote(quoted)
		if err != nil {
			return "", status.Errorf(codes.InvalidArgument,
				"expected shell command: %s to be of the form: %s -c \"<command>\"."+
					" Failed to unquote command with error %s",
				cmd, splitCmd[0], err)
		}
		splitCmd = append(splitCmd[:2], unquoted)
	}

	config := &container.Config{
		Cmd:          splitCmd,
		Labels:       optionz.Labels,
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

	// Handle Network
	if optionz.Network != "" {
		hostConfig.NetworkMode = container.NetworkMode(optionz.Network)
	}

	// Handle Capabilities
	if optionz.Capabilities != nil {
		caps := optionz.Capabilities.(*cpb.StartContainerRequest_Capabilities)
		hostConfig.CapAdd = caps.GetAdd()
		hostConfig.CapDrop = caps.GetRemove()
	}

	// Handle RestartPolicy
	if optionz.RestartPolicy != nil {
		restartPolicy := optionz.RestartPolicy.(*cpb.StartContainerRequest_Restart)

		var policy container.RestartPolicyMode
		switch restartPolicy.GetPolicy() {
		case cpb.StartContainerRequest_Restart_ALWAYS:
			policy = container.RestartPolicyAlways
		case cpb.StartContainerRequest_Restart_ON_FAILURE:
			policy = container.RestartPolicyOnFailure
		case cpb.StartContainerRequest_Restart_NONE:
			policy = container.RestartPolicyDisabled
		case cpb.StartContainerRequest_Restart_UNLESS_STOPPED:
			policy = container.RestartPolicyUnlessStopped
		default:
			return "", status.Errorf(codes.FailedPrecondition, "unkown restart policy '%v'", restartPolicy.GetPolicy())
		}

		hostConfig.RestartPolicy = container.RestartPolicy{
			Name:              policy,
			MaximumRetryCount: int(restartPolicy.GetAttempts()),
		}
	}

	// Handle RunAs
	if optionz.RunAs != nil {
		runAs := optionz.RunAs.(*cpb.StartContainerRequest_RunAs)
		user := runAs.GetUser()
		if user == "" {
			return "", status.Errorf(codes.FailedPrecondition, "user can not be empty in RunAs option")
		}
		if runAs.GetGroup() != "" {
			user = fmt.Sprintf("%s:%s", user, runAs.GetGroup())
		}
		config.User = user
	}

	resp, err := m.client.ContainerCreate(ctx, config, hostConfig, &network.NetworkingConfig{}, nil, optionz.InstanceName)
	if err != nil {
		return "", status.Errorf(codes.Internal, "unable to create container: %v", err)
	}

	if err := m.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
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
