package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/moby/moby/pkg/jsonmessage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
)

// ImagePush pushes the container file to the containerz server. It can optionally tag the
// container, otherwise it will use the name and tag provided in the file.
func (m *Manager) ImagePush(ctx context.Context, file *os.File, opts ...options.Option) (string, string, error) {
	if file == nil {
		return "", "", status.Error(codes.InvalidArgument, "file must be supplied")
	}

	options := options.ApplyOptions(opts...)

	resp, err := m.client.ImageLoad(ctx, file, true)
	if err != nil {
		return "", "", status.Errorf(codes.Internal, "unable to load image: %v", err)
	}
	defer resp.Body.Close()

	if resp.Body != nil && resp.JSON {
		var jm jsonmessage.JSONMessage
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&jm); err != nil {
			return "", "", status.Convert(err).Err()
		}

		imageNameAndTag := extractImageNameFromStream(jm.Stream)
		if options.TargetName != "" && imageNameAndTag != "" {
			if err := m.client.ImageTag(ctx, imageNameAndTag, fmt.Sprintf("%s:%s", options.TargetName, options.TargetTag)); err != nil {
				return "", "", status.Convert(err).Err()
			}
			return options.TargetName, options.TargetTag, nil
		}

		parts := strings.SplitN(imageNameAndTag, ":", 2)
		return parts[0], parts[1], nil
	}

	return options.TargetName, options.TargetTag, nil
}

func extractImageNameFromStream(stream string) string {
	if len(stream) == 0 {
		return ""
	}

	return strings.Trim(strings.Replace(stream, "Loaded image: ", "", 1), "\n")
}
