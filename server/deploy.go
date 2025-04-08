// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/openconfig/containerz/chunker"
	"github.com/openconfig/containerz/containers"
	cpb "github.com/openconfig/gnoi/containerz"
	"k8s.io/klog/v2"
)

// pluginLocation is the location where plugins are expected to be written to.
var pluginLocation = "/plugins"

// Deploy sets a container image on the target. The container is sent as
// a sequential stream of messages containing up to 64KB of data. Upon
// reception of a valid container, the target must load it into its registry.
// Whether the registry is local or remote is target and deployment specific.
// A valid container is one that has passed its checksum.
func (s *Server) Deploy(srv cpb.Containerz_DeployServer) error {

	msg, err := srv.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		klog.Errorf("deploy failed: %v", err)
		return status.Errorf(codes.Internal, "%v", err)
	}

	switch req := msg.GetRequest().(type) {
	case *cpb.DeployRequest_Content, *cpb.DeployRequest_ImageTransferEnd:
		return status.Error(codes.Unavailable, "must send send a TransferImage message first")
	case *cpb.DeployRequest_ImageTransfer:
		if req.ImageTransfer.GetRemoteDownload() != nil {
			opts := []options.Option{
				options.WithStream(srv),
				options.WithRegistryAuth(req.ImageTransfer.GetRemoteDownload().GetCredentials()),
			}
			if err := s.mgr.ImagePull(srv.Context(), req.ImageTransfer.GetName(), req.ImageTransfer.GetTag(), opts...); err != nil {
				return err
			}

			return srv.Send(&cpb.DeployResponse{
				Response: &cpb.DeployResponse_ImageTransferSuccess{
					ImageTransferSuccess: &cpb.ImageTransferSuccess{
						Name: req.ImageTransfer.GetName(),
						Tag:  req.ImageTransfer.GetTag(),
					},
				},
			})
		}

		return s.handleImageTransfer(srv.Context(), srv, req.ImageTransfer)
	default:
		return status.Errorf(codes.InvalidArgument, "unknown request type %T", msg.GetRequest())
	}

}

func (s *Server) handleImageTransfer(ctx context.Context, srv cpb.Containerz_DeployServer, transfer *cpb.ImageTransfer) error {
	if err := checkDiskSpace(s.tmpLocation, transfer.GetImageSize()); err != nil {
		return err
	}

	chunkWriter, err := chunker.NewWriter(s.tmpLocation, s.chunkSize)
	if err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	if err := srv.Send(&cpb.DeployResponse{
		Response: &cpb.DeployResponse_ImageTransferReady{
			ImageTransferReady: &cpb.ImageTransferReady{
				ChunkSize: int32(s.chunkSize),
			},
		},
	}); err != nil {
		return status.Errorf(codes.Unavailable, "client is not ready: %v", err)
	}

	for {
		msg, err := srv.Recv()
		if err == io.EOF {
			return status.Errorf(codes.Unknown, "unexpected EOF while receiving image: %v", err)
		}

		switch req := msg.GetRequest().(type) {
		case *cpb.DeployRequest_Content:
			if _, err := chunkWriter.Write(req.Content); err != nil {
				return err
			}

			if chunkWriter.Size() > transfer.GetImageSize() {
				return status.Errorf(codes.InvalidArgument, "too much data received")
			}

			if err := srv.Send(&cpb.DeployResponse{
				Response: &cpb.DeployResponse_ImageTransferProgress{
					ImageTransferProgress: &cpb.ImageTransferProgress{
						BytesReceived: chunkWriter.Size(),
					},
				},
			}); err != nil {
				return status.Errorf(codes.Unavailable, "client is not ready: %v", err)
			}

		case *cpb.DeployRequest_ImageTransferEnd:
			if transfer.IsPlugin {
				if err := moveFile(chunkWriter.File().Name(), filepath.Join(pluginLocation, fmt.Sprintf("%s.tar", transfer.GetName()))); err != nil {
					return status.Errorf(codes.Internal, "unable to move plugin: %v", err)
				}

				if err := srv.Send(&cpb.DeployResponse{
					Response: &cpb.DeployResponse_ImageTransferSuccess{
						ImageTransferSuccess: &cpb.ImageTransferSuccess{
							Name:      transfer.GetName(),
							ImageSize: chunkWriter.Size(),
						},
					},
				}); err != nil {
					return status.Errorf(codes.Unavailable, "client is not ready: %v", err)
				}
				return nil
			}

			image, tag, err := s.mgr.ImagePush(ctx, chunkWriter.File(), options.WithTarget(transfer.GetName(), transfer.GetTag()))
			if err != nil {
				return err
			}

			if err := srv.Send(&cpb.DeployResponse{
				Response: &cpb.DeployResponse_ImageTransferSuccess{
					ImageTransferSuccess: &cpb.ImageTransferSuccess{
						Name:      image,
						Tag:       tag,
						ImageSize: chunkWriter.Size(),
					},
				},
			}); err != nil {
				return status.Errorf(codes.Unavailable, "client is not ready: %v", err)
			}

			return nil
		default:
			return status.Errorf(codes.Internal, "unexpected message type %T", msg.GetRequest())
		}
	}
}

func checkDiskSpace(loc string, bytesNeeded uint64) error {
	availableSpace, err := diskSpace(loc)
	if err != nil {
		return status.Errorf(codes.Internal, "unable to check free space: %v", err)
	}

	if availableSpace < bytesNeeded {
		return status.Error(codes.ResourceExhausted, "not enough space to store image")
	}

	return nil
}

func diskSpace(loc string) (uint64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(loc, &stat); err != nil {
		return 0, err
	}

	// Available blocks * size per block = available space in bytes
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}

// moveFile moves a file by copying it and deleting the source. This is needed because os.Rename
// only works within one device (i.e. mountpoint). The replication server's temp location and
// actual location may be on different devices.
func moveFile(sourcePath, destPath string) error {
	// idempotent create of the destPath directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create %s with error %s",
			filepath.Dir(destPath))
	}
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("unable to open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("unable to open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	// Success, now delete the original file.
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("failed removing original file: %s", err)
	}
	return nil
}
