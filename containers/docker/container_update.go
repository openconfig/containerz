package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	"github.com/openconfig/containerz/containers"
)

type instanceConfig struct {
	config     *container.Config
	hostConfig *container.HostConfig
}

func (m *Manager) jsonState(ctx context.Context, instance string, cnts []types.Container) (types.ContainerJSON, error) {
	// Get Container ID
	var cntID string
	for _, cnt := range cnts {
		if containerMatchesInstance(cnt, instance) {
			cntID = cnt.ID
			break
		}
	}
	if cntID == "" {
		return types.ContainerJSON{}, status.Errorf(codes.NotFound, "instance name %s not found", instance)
	}

	// Fetch instance configs.
	cntJSON, err := m.client.ContainerInspect(ctx, cntID)
	if err != nil {
		return types.ContainerJSON{}, status.Errorf(codes.Unknown, "failed to inspect container %s: %v", cntID, err)
	}

	return cntJSON, nil
}

func (m *Manager) performContainerUpdate(ctx context.Context, instance, image, tag, cmd string, cnts []types.Container, opts ...options.Option) (string, error) {
	// Don't forget to notify the manager that update for this instance has finished.
	defer func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.updateInProgress, instance)
	}()

	// Save the current config in case we need to fallback.
	oldCntJSON, err := m.jsonState(ctx, instance, cnts)
	if err != nil {
		return "", err
	}

	if err := m.ContainerStop(ctx, instance, opts...); err != nil {
		// If the container stop fails, there shouldn't be any changes to restore.
		return "", status.Errorf(codes.Internal, "failed update of instance %s due to: %v", instance, err)
	}
	// ContainerStop will stop the container - we want to additionally remove this instance here.
	if err := m.client.ContainerRemove(ctx, instance, container.RemoveOptions{}); err != nil {
		return "", status.Errorf(codes.Internal, "failed update of instance %s due to: %v", instance, err)
	}

	// Attempting to create & start a container with the new config.
	opts = append(opts, options.WithInstanceName(instance))
	if _, err = m.ContainerStart(ctx, image, tag, cmd, opts...); err == nil { // if NO error
		return instance, nil
	}

	// There was some error, let's try to restore previous state.
	errPfx := fmt.Sprintf("failed to update instance %s due to: %v", instance, err)

	resp, err := m.client.ContainerCreate(ctx, oldCntJSON.Config, oldCntJSON.HostConfig, &network.NetworkingConfig{}, nil, instance)
	if err != nil {
		return "", status.Errorf(codes.Internal, "%s; restoration of previous state failed when creating container: %v", errPfx, err)
	}

	if err := m.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", status.Errorf(codes.Internal, "%s; restoration of previous state failed when starting container: %v", errPfx, err)
	}

	return instance, status.Errorf(codes.Internal, "%s; yet, restoration of previous state succeeded", errPfx)
}

func (m *Manager) performContainerUpdatePrechecks(ctx context.Context, instance, image, tag, cmd string, async bool, opts ...options.Option) ([]types.Container, error) {
	optionz := options.ApplyOptions(opts...)

	// Get available images and containers to perform checks on.
	cnts, err := m.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	images, err := m.client.ImageList(ctx, imagetypes.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	// Ensure that image exists.
	ref := fmt.Sprintf("%s:%s", image, tag)
	if err := findImage(ref, images); err != nil {
		return nil, err
	}

	// Ensure that instance exists.
	if err := checkInstanceExists(instance, cnts); err != nil {
		return nil, err
	}

	// Ensure that the provided port mapping is feasible.
	if err := checkPortAvailability(optionz.PortMapping, cnts, instance); err != nil {
		return nil, err
	}

	return cnts, nil
}

// stageContainerUpdate ensures that this instance does not have an in-progress update running.
func (m *Manager) stageContainerUpdate(instance string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.updateInProgress[instance]; ok {
		return status.Errorf(codes.Unavailable, "container %s is already being updated", instance)
	}

	m.updateInProgress[instance] = struct{}{} // Not updating already, fine to start new update.
	return nil
}

// ContainerUpdate updates a running container to the image specified in the
// request. By default the operation is synchronous which means that the
// request will only return once the container has either been successfully
// updated or the update has failed. If the client requests an asynchronous
// update then the server must perform all validations (e.g. does the
// requested image exist on the system or does the instance name exist) and
// return to the client and the update happens asynchronously. It is up to the
// client to check if the update actually updates the container to the
// requested version or not.
// In both synchronous and asynchronous mode, the update process is a
// break-before-make process as resources bound to the old container must be
// released prior to launching the new container.
// If the update fails, the server must restore the previous version of the
// container. This can either be a start of the previous container or by
// starting a new container with the old image.
// It must use the provided StartContainerRequest provided in the
// params field.
// If a container exists but is not running should still upgrade the container
// and start it.
// The client should only depend on the client being restarted. Any ephemeral
// state (date written to memory or the filesystem) cannot be depended upon.
// In particular, the contents of the filesystem are not guaranteed during a
// rollback.
func (m *Manager) ContainerUpdate(ctx context.Context, instance, image, tag, cmd string, async bool, opts ...options.Option) (string, error) {

	// Perform all pre-update checks.
	cnts, err := m.performContainerUpdatePrechecks(ctx, instance, image, tag, cmd, async, opts...)
	if err != nil {
		return "", err
	}

	// Ensure that this instance does not have an in-progress update running & stage the update.
	if err := m.stageContainerUpdate(instance); err != nil {
		return "", err
	}

	// All checks passed, proceed to the actual (synchronous or asynchronous) update.
	if async {
		klog.Infof("Starting asynchronous update of instance %s to image %s:%s with cmd %s and options %+v", instance, image, tag, cmd, opts)
		deadline, ok := ctx.Deadline()
		// Override the cancellation from the parent context.
		// This allows the RPC to exit while the async update completes.
		// As the parent ctx can no longer be cancelled, the deadline/timeout will serve as the
		// only way to timeout the aysnc update.
		ctx = context.WithoutCancel(ctx)
		var cancel context.CancelFunc
		if ok {
			// If the parent context had a deadline, then this deadline should be maintained.
			ctx, cancel = context.WithDeadline(ctx, deadline)
		} else {
			// If the parent context had no deadline, then set a deadline of 5 minutes as default.
			ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
		}
		go func() {
			defer cancel()
			// There can only be one go routine per instance name due to the mutex handling.
			updatedInstance, err := m.performContainerUpdate(
				ctx, instance, image, tag, cmd, cnts, opts...)
			if err != nil {
				klog.Infof("Async container update failed. Error is %s", err)
				return
			}
			klog.Infof("Async container update successful. Updated container is %q",
				updatedInstance)
		}()
		return instance, nil
	}
	return m.performContainerUpdate(ctx, instance, image, tag, cmd, cnts, opts...)
}

// checkInstanceExists checks whether a container with the given instance name exists.
func checkInstanceExists(instance string, cnts []types.Container) error {
	for _, cnt := range cnts {
		for _, name := range cnt.Names {
			strippedname := strings.Replace(name, "/", "", 1)
			if strippedname == instance {
				return nil
			}
		}
	}
	return status.Errorf(codes.NotFound, "instance name %s not found", instance)
}

// checkPortAvailability checks whether the provided port map relies on in-use ports.
// Notably, this check ignores ports on containers matching the provided ignoreInstance name.
func checkPortAvailability(ports map[uint32]uint32, cnts []types.Container, ignoreInstance string) error {
	for _, cnt := range cnts {
		// Shall we ignore this container's ports?
		if containerMatchesInstance(cnt, ignoreInstance) {
			continue
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

// Checks whether any of the container's names matches the instance name.
func containerMatchesInstance(cnt types.Container, instance string) bool {
	for _, name := range cnt.Names {
		strippedname := strings.Replace(name, "/", "", 1)
		if strippedname == instance {
			return true
		}
	}
	return false
}
