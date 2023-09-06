package docker

import (
	"context"
	"time"

	"github.com/moby/moby/v/v24/api/types/container"
	"github.com/moby/moby/v/v24/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
)

// ContainerStop stops a container. If the Force option is set and a timeout
// is specified in the context, the contains if forcefully terminated after that timeout.
// If the Force option is set but no timeout is provided the container's StopTimeout
// value is used, if set, otherwise the engine default.
// If the Force option is not set, no forceful termination is performed.
func (m *Manager) ContainerStop(ctx context.Context, instance string, opts ...options.ImageOption) error {
	optionz := options.ApplyOptions(opts...)

	cnts, err := m.client.ContainerList(ctx, types.ContainerListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return err
	}

	// check if the container exists.
	if err := checkExistingInstanceAndPorts(instance, nil, cnts); err == nil {
		return status.Errorf(codes.NotFound, "container %s was not found", instance)
	}

	// a negative timeout indicates to docker that no forceful termination should
	// occur.
	duration := -1
	if optionz.Force {
		// compute duration from context deadline
		duration = 0
		timeoutTime, ok := ctx.Deadline()
		if ok {
			duration = int(timeoutTime.Sub(time.Now()).Seconds())
		}
	}

	pDuration := &duration
	if duration == 0 {
		pDuration = nil
	}

	if err := m.client.ContainerStop(ctx, instance, container.StopOptions{Timeout: pDuration}); err != nil {
		return err
	}

	if err := m.client.ContainerRemove(ctx, instance, types.ContainerRemoveOptions{
		Force: optionz.Force,
	}); err != nil {
		return err
	}

	return nil
}
