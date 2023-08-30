package client

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	cpb "github.com/openconfig/gnoi/containerz"
)

var (

	// ErrNotFound indicates that the specified image was not found on the target system.
	ErrNotFound = status.Error(codes.NotFound, "resource was not found")

	// ErrRunning indicates that there is a container running this image.
	ErrRunning = status.Error(codes.FailedPrecondition, "resource is running")
)

// Remove removes an image from the target system. It returns nil upon success. Otherwise it
// returns an error indicating whether the image was not found or is associated to running
// container.
func (c *Client) Remove(ctx context.Context, image string, tag string) error {
	resp, err := c.cli.Remove(ctx, &cpb.RemoveRequest{
		Name: image,
		Tag:  tag,
	})
	if err != nil {
		return err
	}

	switch resp.GetCode() {
	case cpb.RemoveResponse_SUCCESS:
		return nil
	case cpb.RemoveResponse_NOT_FOUND:
		return ErrNotFound
	case cpb.RemoveResponse_RUNNING:
		return ErrRunning
	default:
		return errors.New("unknwon error occurred")
	}
}
