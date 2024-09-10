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

package client

import (
	"context"
	"io"
        "fmt"
        "strings"
	"k8s.io/klog/v2"
	cpb "github.com/openconfig/gnoi/containerz"
)

// ListImage implements the client logic for listing the existing containers on the target system.
func (c *Client) ListImage(ctx context.Context, limit int32, filter []string) (<-chan *ImageInfo, error) {
       	imgFilters, err := imgFilters(filter)
	if err != nil {
		return nil, err
	} 
	req := &cpb.ListImageRequest{
		Limit:  limit,
		Filter: imgFilters,
	}

	dcli, err := c.cli.ListImage(ctx, req)
	if err != nil {
		return nil, err
	}

	ch := make(chan *ImageInfo, 100)
	go func() {
		defer dcli.CloseSend()
		defer close(ch)
		for {
			msg, err := dcli.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				nonBlockingChannelSend(ctx, ch, &ImageInfo{
					Error: err,
				})
				return
			}

			if nonBlockingChannelSend(ctx, ch, &ImageInfo{
				ID:        msg.GetId(),
				ImageName: msg.GetImageName(),
				Tag:     msg.GetTag(),
			}) {
				klog.Warningf("operation cancelled; returning")
				return
			}

		}
	}()

	return ch, nil
}
func imgFilters (filters []string) ([]*cpb.ListImageRequest_Filter, error) {
	mapping := make([]*cpb.ListImageRequest_Filter, 0, len(filters))
	for _, f := range filters {
		parts := strings.Split(f, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid filter: %s", f)
		}
                values := strings.Split(parts[1], ",")
		mapping = append(mapping, &cpb.ListImageRequest_Filter{
			Key:   parts[0],
			Value: values,
		})
	}
	return mapping, nil
}
