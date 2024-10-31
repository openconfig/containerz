package docker

import (
	"context"
	"io"
	"strings"

	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/filters/filters"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/types"
	"github.com/openconfig/containerz/containers"

	cpb "github.com/openconfig/gnoi/containerz"
)

// ImageList lists the images present on the target.
func (m *Manager) ImageList(ctx context.Context, all bool, limit int32, srv options.ListImageStreamer, opts ...options.Option) error {
	optionz := options.ApplyOptions(opts...)
	// ImageListOptions doesn't support a limit. We add one artificially below.
	imgOpts := types.ImageListOptions{All: all}

	var kvPairs []filters.KeyValuePair
	for key, values := range optionz.Filter {
		for _, value := range values {
			kvPairs = append(kvPairs, filters.KeyValuePair{Key: string(key), Value: value})
		}
	}
	imgOpts.Filters = filters.NewArgs(kvPairs...)

	images, err := m.client.ImageList(ctx, imgOpts)
	if err != nil {
		return err
	}

	// Artificially limit the response.
	if limit > 0 && limit < int32(len(images)) {
		images = images[:limit]
	}

	for _, image := range images {
		if err := srv.Send(imageToResponse(image)); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}

	return nil
}

func imageToResponse(image types.ImageSummary) *cpb.ListImageResponse {
	var name string
	var tags []string
	for _, tag := range image.RepoTags {
		parts := strings.SplitN(tag, ":", 2)
		if name == "" {
			name = parts[0] // This should be the same for each tag.
		}
		if len(parts) > 1 {
			tags = append(tags, parts[1])
		}
	}

	return &cpb.ListImageResponse{
		Id:        image.ID,
		ImageName: name,
		Tag:       strings.Join(tags, ","),
	}
}
