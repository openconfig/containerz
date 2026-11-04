package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/client"
	"github.com/moby/moby/pkg/jsonmessage"
	"github.com/openconfig/containerz/containers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ImagePush pushes the container file to the containerz server. It can optionally tag the
// container, otherwise it will use the name and tag provided in the file.
func (m *Manager) ImagePush(ctx context.Context, file *os.File, opts ...options.Option) (string, string, error) {
	if file == nil {
		return "", "", status.Error(codes.InvalidArgument, "file must be supplied")
	}

	options := options.ApplyOptions(opts...)

	resp, err := m.client.ImageLoad(ctx, file, client.ImageLoadWithQuiet(true))
	if err != nil {
		return "", "", status.Errorf(codes.Internal, "unable to load image: %v", err)
	}
	defer resp.Body.Close()

	if resp.Body != nil && resp.JSON {
		loaded, err := loadedImageFromStream(json.NewDecoder(resp.Body))
		if err != nil {
			return "", "", err
		}

		if options.TargetName != "" {
			if err := m.client.ImageTag(ctx, loaded.ref(), fmt.Sprintf("%s:%s", options.TargetName, options.TargetTag)); err != nil {
				return "", "", status.Convert(err).Err()
			}
			return options.TargetName, options.TargetTag, nil
		}

		if loaded.name == "" {
			return "", "", status.Errorf(codes.InvalidArgument,
				"image was loaded as %s without a name and tag; a target name must be provided", loaded.id)
		}
		return loaded.name, loaded.tag, nil
	}

	return options.TargetName, options.TargetTag, nil
}

const (
	// loadedImagePrefix is the prefix docker uses to report a loaded image that
	// has a name and tag.
	loadedImagePrefix = "Loaded image: "
	// loadedImageIDPrefix is the prefix docker uses to report a loaded image
	// that has no name or tag.
	loadedImageIDPrefix = "Loaded image ID: "
)

// loadedImage describes the image reported by docker after an image load.
// Either name (and tag) or id is set.
type loadedImage struct {
	name, tag string
	id        string
}

// ref returns a reference to the loaded image that can be used as the source
// of an image tag operation.
func (l loadedImage) ref() string {
	if l.name == "" {
		return l.id
	}
	return l.name + ":" + l.tag
}

// loadedImageFromStream reads all messages docker sends in response to an
// image load and returns the first image it reports as loaded. Other messages,
// such as the warning that an existing image with the same tag was renamed,
// are ignored. Errors reported in the stream are returned.
func loadedImageFromStream(dec *json.Decoder) (loadedImage, error) {
	var loaded loadedImage
	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return loadedImage{}, status.Errorf(codes.Internal, "unable to decode image load response: %v", err)
		}

		if jm.Error != nil {
			return loadedImage{}, status.Errorf(codes.Internal, "unable to load image: %s", jm.Error.Message)
		}
		if jm.ErrorMessage != "" {
			return loadedImage{}, status.Errorf(codes.Internal, "unable to load image: %s", jm.ErrorMessage)
		}

		if loaded.name != "" || loaded.id != "" {
			// Keep reading to surface errors reported after the image was loaded.
			continue
		}
		for _, line := range strings.Split(jm.Stream, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, loadedImageIDPrefix) {
				loaded.id = strings.TrimPrefix(line, loadedImageIDPrefix)
				break
			}
			if strings.HasPrefix(line, loadedImagePrefix) {
				loaded.name, loaded.tag = splitImageReference(strings.TrimPrefix(line, loadedImagePrefix))
				break
			}
		}
	}

	if loaded.name == "" && loaded.id == "" {
		return loadedImage{}, status.Error(codes.Internal, "image load response did not report a loaded image")
	}
	return loaded, nil
}

// splitImageReference splits an image reference of the form name[:tag] into
// its name and tag. The tag defaults to "latest". A colon that is part of a
// registry host:port is not treated as the tag separator.
func splitImageReference(ref string) (string, string) {
	i := strings.LastIndex(ref, ":")
	if i < 0 || strings.Contains(ref[i+1:], "/") {
		return ref, "latest"
	}
	return ref[:i], ref[i+1:]
}
