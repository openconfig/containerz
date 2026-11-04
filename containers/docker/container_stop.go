package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"github.com/openconfig/containerz/containers"
)

// maximumStopTimeout sets a cap on how long the docker can wait before
// sending a SIGKILL after the initial SIGTERM.
// 10 (seconds) is chosen as this is the default timeout used by container.StopOptions
const maximumStopTimeout = 10

// minimumStopTimeout is the lower bound (in seconds) for the timeout that is
// derived from the context deadline. It gives the container at least some time
// to handle SIGTERM, and avoids passing 0 (immediate SIGKILL) or a negative
// value (no SIGKILL at all) to docker when the deadline is close or has passed.
const minimumStopTimeout = 1

// forcedStopTimeout returns the number of seconds docker should wait after
// SIGTERM before sending SIGKILL, given the time remaining until the RPC
// deadline. It uses half of the remaining time, so that the RPC context has not
// expired when ContainerStop returns, bounded by [minimumStopTimeout,
// maximumStopTimeout].
func forcedStopTimeout(remaining time.Duration) int {
	seconds := int((remaining / 2).Round(time.Second) / time.Second)
	return max(minimumStopTimeout, min(seconds, maximumStopTimeout))
}

// ContainerStop stops a container. If the Force option is set and a timeout
// is specified in the context, the contains if forcefully terminated after that timeout.
// If the Force option is set but no timeout is provided the container's StopTimeout
// value is used, if set, otherwise the engine default.
// If the Force option is not set, no forceful termination is performed.
// If the Restart option is set, the container is started again after it has
// stopped, using the same termination semantics.
func (m *Manager) ContainerStop(ctx context.Context, instance string, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)

	cnts, err := m.client.ContainerList(ctx, container.ListOptions{
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
	// occur. A nil timeout makes docker use the container's StopTimeout, if set,
	// otherwise the engine default.
	duration := -1
	pDuration := &duration
	if optionz.Force {
		pDuration = nil
		if timeoutTime, ok := ctx.Deadline(); ok {
			duration = forcedStopTimeout(time.Until(timeoutTime))
			pDuration = &duration
		}
	}

	if optionz.Restart {
		// ContainerRestart stops the container using the same timeout semantics
		// as ContainerStop and then starts it again. The container keeps its ID,
		// configuration and restart policy.
		if err := m.client.ContainerRestart(ctx, instance, container.StopOptions{Timeout: pDuration}); err != nil {
			klog.Warningf("container %s failed to restart", instance)
			return status.Errorf(codes.Unknown, "failed to restart container %s with error %s",
				instance, err)
		}
		return nil
	}

	if err := m.client.ContainerStop(ctx, instance, container.StopOptions{Timeout: pDuration}); err != nil {
		klog.Warningf("container %s failed to stop", instance)
		return status.Errorf(codes.Unknown, "failed to stop container %s with error %s",
			instance, err)
	}

	return nil
}
