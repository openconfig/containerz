package server

import (
	"google3/third_party/openconfig/containerz/containers/options"
	cpb "github.com/openconfig/gnoi/containerz"
)

// Log streams the logs of a running container. If the container if no longer
// running this operation streams the latest logs and returns.
func (s *Server) Log(request *cpb.LogRequest, srv cpb.Containerz_LogServer) error {
	opts := []options.ImageOption{}
	if request.GetFollow() {
		opts = append(opts, options.Follow())
	}

	return s.mgr.ContainerLogs(srv.Context(), request.GetInstanceName(), srv, opts...)
}
