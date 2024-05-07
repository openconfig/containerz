package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types"
	"github.com/moby/moby/pkg/jsonmessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
	tpb "github.com/openconfig/gnoi/types"
)

// ContainerPull pull a container from a registry to this containerz server. Based on the options
// specified  it can tag the container, stream responses to the client, and perform registry
// authentication.
func (m *Manager) ContainerPull(ctx context.Context, image, tag string, opts ...options.Option) error {
	switch {
	case image == "":
		return status.Error(codes.InvalidArgument, "an image name must be supplied.")
	case tag == "":
		tag = "latest"
	}

	options := options.ApplyOptions(opts...)

	auth, err := m.registryLogin(options.Credentials)
	if err != nil {
		return err
	}

	resp, err := m.client.ImagePull(ctx, fmt.Sprintf("%s:%s", image, tag), types.ImagePullOptions{
		RegistryAuth: auth.IdentityToken,
	})
	if err != nil {
		return status.Errorf(codes.Internal, "unable to pull container: %v", err)
	}
	defer resp.Close()

	if options.StreamClient != nil {
		if err := streamOutput(options.StreamClient, resp); err != nil {
			// TODO(alshabib): should we ignore this error and only log it?
			return err
		}
	}

	if options.TargetName != "" && options.TargetTag != "" {
		if err := m.client.ImageTag(ctx, fmt.Sprintf("%s:%s", image, tag), fmt.Sprintf("%s:%s", options.TargetName, options.TargetTag)); err != nil {
			return status.Errorf(codes.Internal, "unable to tag container: %v", err)
		}
	}

	return nil
}

func (m *Manager) registryLogin(creds *tpb.Credentials) (registry.AuthenticateOKBody, error) {
	if creds == nil {
		return registry.AuthenticateOKBody{}, nil
	}

	return registry.AuthenticateOKBody{}, status.Error(codes.Unimplemented, "registry auth not yet implemented")
}

func streamOutput(srv options.Stream, resp io.ReadCloser) error {
	dec := json.NewDecoder(resp)

	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if jm.Progress == nil || jm.Progress.Current == 0 {
			continue
		}

		if err := srv.Send(&cpb.DeployResponse{
			Response: &cpb.DeployResponse_ImageTransferProgress{
				ImageTransferProgress: &cpb.ImageTransferProgress{
					BytesReceived: uint64(jm.Progress.Current),
				},
			},
		}); err != nil {
			return err
		}
	}

	return nil
}
