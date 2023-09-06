package docker

import (
	"context"
	"io"
	"strings"

	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/filters/filters"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/types"
	"/containers/options"
	cpb "github.com/openconfig/gnoi/containerz"
)

// ContainerList lists the containers present on the target.
func (m *Manager) ContainerList(ctx context.Context, all bool, limit int32, srv options.ListStreamer, opts ...options.ImageOption) error {
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
		if err := srv.Send(&cpb.ListResponse{
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

func stringToStatus(state string) cpb.ListResponse_Status {
	switch {
	case strings.Contains(state, "Up"):
		return cpb.ListResponse_RUNNING
	case strings.Contains(state, "Exited"):
		return cpb.ListResponse_STOPPED
	default:
		return cpb.ListResponse_UNSPECIFIED
	}
}
