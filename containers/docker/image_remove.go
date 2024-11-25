package docker

import (
	"context"

	"fmt"

	"github.com/docker/docker/api/types/container"
	imagetypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
)

// ImageRemove removes an image provided it is not related to a running container. Otherwise,
// it returns an error.
func (m *Manager) ImageRemove(ctx context.Context, image, tag string, opts ...options.Option) error {
	option := options.ApplyOptions(opts...)

	images, err := m.client.ImageList(ctx, imagetypes.ListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return err
	}

	ref := fmt.Sprintf("%s:%s", image, tag)
	if err := findImage(ref, images); err != nil {
		return err
	}

	cnts, err := m.client.ContainerList(ctx, container.ListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return err
	}

	state := findImageFromContainer(ref, cnts)
	if state != nil {
		if option.Force {
			_, err := m.client.ImageRemove(ctx, ref, imagetypes.RemoveOptions{
				Force: option.Force,
			})
			return err
		}
		return state.Err()
	}

	_, err = m.client.ImageRemove(ctx, ref, imagetypes.RemoveOptions{})
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

func findImage(ref string, summaries []imagetypes.Summary) error {
	for _, summary := range summaries {
		for _, name := range summary.RepoTags {
			if ref == name {
				return nil
			}
		}
	}

	return status.Errorf(codes.NotFound, "image %s not found", ref)
}
