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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"
	cpb "github.com/openconfig/gnoi/containerz"
)

type fakeLoggingContainerzServer struct {
	fakeContainerzServer

	sendMsgs         []*cpb.LogResponse
	receivedMessages *cpb.LogRequest
	err              error
}

func (f *fakeLoggingContainerzServer) Log(req *cpb.LogRequest, srv cpb.Containerz_LogServer) error {
	f.receivedMessages = req

	if f.err != nil {
		return f.err
	}

	for _, resp := range f.sendMsgs {
		if err := srv.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func TestLogs(t *testing.T) {
	tests := []struct {
		name       string
		inInstance string
		inFollow   bool
		inErr      error
		inMsgs     []*cpb.LogResponse

		wantLogs []*LogMessage
		wantMsg  *cpb.LogRequest
	}{
		{
			name:       "simple",
			inInstance: "some-instance",
			inFollow:   true,
			inMsgs: []*cpb.LogResponse{
				&cpb.LogResponse{
					Msg: "logs",
				},
				&cpb.LogResponse{
					Msg: "more-logs",
				},
			},
			wantLogs: []*LogMessage{
				&LogMessage{
					Msg: "logs",
				},
				&LogMessage{
					Msg: "more-logs",
				},
			},
			wantMsg: &cpb.LogRequest{
				InstanceName: "some-instance",
				Follow:       true,
			},
		},
		{
			name:       "no-such-container",
			inInstance: "some-instance",
			inFollow:   true,
			inErr:      status.Error(codes.NotFound, "container some-instance not found"),
			wantLogs: []*LogMessage{
				&LogMessage{
					Error: status.Error(codes.NotFound, "container some-instance not found"),
				},
			},
			wantMsg: &cpb.LogRequest{
				InstanceName: "some-instance",
				Follow:       true,
			},
		},
	}

	ctx := context.Background()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fcm := &fakeLoggingContainerzServer{
				err:      tc.inErr,
				sendMsgs: tc.inMsgs,
			}
			addr, stop := newServer(t, fcm)
			defer stop()
			cli, err := NewClient(ctx, addr)
			if err != nil {
				t.Fatalf("NewClient(%v) returned an unexpected error: %v", addr, err)
			}

			doneCh := make(chan struct{})
			got := []*LogMessage{}

			ch, err := cli.Logs(ctx, tc.inInstance, tc.inFollow)
			if err != nil {
				t.Fatalf("Logs(%s, %t) returned an unexpected error: %v", tc.inInstance, tc.inFollow, err)
			}

			go func() {
				for log := range ch {
					got = append(got, log)
				}
				close(doneCh)
			}()
			<-doneCh

			if diff := cmp.Diff(tc.wantLogs, got, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Logs(%s, %t)returned an unexpected diff (-want +got):\n%s", tc.inInstance, tc.inFollow, diff)
			}

			if diff := cmp.Diff(tc.wantMsg, fcm.receivedMessages, protocmp.Transform()); diff != "" {
				t.Errorf("Logs(%s, %t)returned an unexpected diff (-want +got):\n %s", tc.inInstance, tc.inFollow, diff)
			}
		})
	}
}
