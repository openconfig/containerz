package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
)

// Remove deletes containers that match the spec defined in the request. If
// the specified container does not exist, this operation is a no-op.
func (s *Server) Remove(ctx context.Context, request *cpb.RemoveRequest) (*cpb.RemoveResponse, error) {
	// TODO(alshabib: add force to proto)
	if err := s.mgr.ContainerRemove(ctx, request.GetName(), request.GetTag(), options.Force()); err != nil {
		stErr, ok := status.FromError(err)
		if !ok {
			return nil, status.Errorf(codes.Internal, "unknown containerz state: %v", err)
		}

		switch stErr.Code() {
		case codes.NotFound:
			return &cpb.RemoveResponse{
				Code:   cpb.RemoveResponse_NOT_FOUND,
				Detail: stErr.Message(),
			}, nil
		case codes.Unavailable:
			return &cpb.RemoveResponse{
				Code:   cpb.RemoveResponse_RUNNING,
				Detail: stErr.Message(),
			}, nil
		}
		return nil, err
	}

	return &cpb.RemoveResponse{
		Code: cpb.RemoveResponse_SUCCESS,
	}, nil
}
