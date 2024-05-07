package docker

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

// ContainerList lists the containers present on the target.
func (m *Manager) ContainerList(ctx context.Context, all bool, limit int32, srv options.ListContainerStreamer, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)

	cntOpts := types.ContainerListOptions{All: all, Limit: int(limit)}

	kvPairs := []filters.KeyValuePair{}
	for key, values := range optionz.Filter {
		for _, value := range values {
			kvPairs = append(kvPairs, filters.KeyValuePair{Key: string(key), Value: value})
		}
	}

	cntOpts.Filters = filters.NewArgs(kvPairs...)

	cnts, err := m.client.ContainerList(ctx, cntOpts)
	if err != nil {
		return err
	}

	for _, cnt := range cnts {
		if err := srv.Send(&cpb.ListContainerResponse{
			Id: cnt.ID,
			// TODO(alshabib): make Name a repeated field.
			Name:      strings.Join(cnt.Names, ","),
			ImageName: cnt.Image,
			Status:    stringToStatus(cnt.Status),
		}); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}

	return nil
}

func stringToStatus(state string) cpb.ListContainerResponse_Status {
	switch {
	case strings.Contains(state, "Up"):
		return cpb.ListContainerResponse_RUNNING
	case strings.Contains(state, "Exited"):
		return cpb.ListContainerResponse_STOPPED
	default:
		return cpb.ListContainerResponse_UNSPECIFIED
	}
}
