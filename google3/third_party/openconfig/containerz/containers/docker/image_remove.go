package docker

import (
	"context"

	"fmt"

	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"/containers/options"
)

// ContainerRemove removes an image provided it is not related to a running container. Otherwise,
// it returns an error.
func (m *Manager) ContainerRemove(ctx context.Context, image, tag string, opts ...options.ImageOption) error {
	option := options.ApplyOptions(opts...)

	images, err := m.client.ImageList(ctx, types.ImageListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return err
	}

	ref := fmt.Sprintf("%s:%s", image, tag)
	if err := findImage(ref, images); err != nil {
		return err
	}

	cnts, err := m.client.ContainerList(ctx, types.ContainerListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return err
	}

	state := findImageFromContainer(ref, cnts)
	if state != nil {
		if option.Force {
			_, err := m.client.ImageRemove(ctx, ref, types.ImageRemoveOptions{
				Force: option.Force,
			})
			return err
		}
		return state.Err()
	}

	_, err = m.client.ImageRemove(ctx, ref, types.ImageRemoveOptions{})
	return err
}

func findImageFromContainer(ref string, cnt []types.Container) *status.Status {
	for _, c := range cnt {
		if c.Image == ref {
			return status.Newf(codes.Unavailable, "image %s has a running container; use force to override", ref)
		}
	}
	return nil
}

func findImage(ref string, summaries []types.ImageSummary) error {
	for _, summary := range summaries {
		for _, name := range summary.RepoTags {
			if ref == name {
				return nil
			}
		}
	}

	return status.Errorf(codes.NotFound, "image %s not found", ref)
}
