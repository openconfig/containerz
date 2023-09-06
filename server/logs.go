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
	"github.com/openconfig/containerz/containers"
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
