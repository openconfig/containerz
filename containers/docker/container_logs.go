package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/moby/moby/v/v24/api/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

// ContainerLogs fetches the logs from a container. It can optionally follow the logs
// and send them back to the client.
func (m *Manager) ContainerLogs(ctx context.Context, instance string, srv options.LogStreamer, opts ...options.ImageOption) error {
	optionz := options.ApplyOptions(opts...)

	cnts, err := m.client.ContainerList(ctx, types.ContainerListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return err
	}

	if len(cnts) == 0 {
		return status.Errorf(codes.NotFound, "container %s not found", instance)
	}

	logOpts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	if optionz.Follow {
		logOpts.Follow = true
	}

	// TODO(alshabib): add this option to proto
	if optionz.Since != 0 {
		logOpts.Since = fmt.Sprintf("%s", optionz.Since)
	}

	// TODO(alshabib): add this option to proto
	if optionz.Until != 0 {
		logOpts.Until = fmt.Sprintf("%s", optionz.Since)
	}

	resp, err := m.client.ContainerLogs(ctx, instance, logOpts)
	if err != nil {
		return err
	}
	defer resp.Close()

	_, err = io.Copy(&logStreamer{srv: srv}, resp)
	return err
}

type logStreamer struct {
	srv options.LogStreamer
}

// Write writes the log message to the provided streamer in logStreamer.
func (l *logStreamer) Write(p []byte) (n int, err error) {
	if err := l.srv.Send(&cpb.LogResponse{
		Msg: string(p),
	}); err != nil {
		if err == io.EOF {
			return 0, io.EOF
		}
		return 0, err
	}

	return len(p), nil
}
