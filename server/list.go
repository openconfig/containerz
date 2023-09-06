package server

import (
	cpb "github.com/openconfig/gnoi/containerz"
)

// List returns all containers that match the spec defined in the request.
func (s *Server) List(request *cpb.ListRequest, srv cpb.Containerz_ListServer) error {
	// TODO (alshabib): fix filter in proto, it needs to be a repeated field
	return s.mgr.ContainerList(srv.Context(), request.GetAll(), request.GetLimit(), srv)
}
