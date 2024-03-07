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

// Package options provides options that can be used by container orchestration implementation.
package options

import (
	"time"

	"google.golang.org/protobuf/proto"
	cpb "github.com/openconfig/gnoi/containerz"
	tpb "github.com/openconfig/gnoi/types"
)

// Stream represents an entity capable of sending responses to the client.
type Stream interface {
	Send(*cpb.DeployResponse) error
}

// LogStreamer is an entity capable of streaming logs.
type LogStreamer interface {
	Send(msg *cpb.LogResponse) error
}

// ListContainerStreamer is an entity capable of streaming container information.
type ListContainerStreamer interface {
	Send(msg *cpb.ListContainerResponse) error
}

// ListVolumeStreamer is an entity capable of streaming volume information.
type ListVolumeStreamer interface {
	Send(msg *cpb.ListVolumeResponse) error
}

// FilterKey represents a key for a filter.
type FilterKey string

const (

	// Image filters for containers based on this image.
	Image FilterKey = "image"

	// Container filters by container name.
	Container = "container"

	// State filters by container state.
	State = "state"

	// Volume filters by volume name.
	Volume = "volume"
)

// Option takes an option and applies it to the set of options when the function is called.
type Option func(*options)

// options contain the supported set of options for pull operations.
type options struct {

	// TargetName image name for the container. If unset, the image will be the name it has in the
	// registry.
	TargetName string

	// TargetTag is the tag to apply to the image. If unset, it will default to "latest".
	TargetTag string

	// Credentials are the set of credentials to use to login to the registry.
	Credentials *tpb.Credentials

	// StreamClient is the client to stream progress reports to.
	StreamClient Stream

	// Force indicates that the container implementation should attempt to force the operation.
	Force bool

	// InstanceName is the name of a running container instance.
	InstanceName string

	// PortMapping is a mapping of internal to external port for a container.
	PortMapping map[uint32]uint32

	// EnvMapping is a set of environment variables to set in the container
	EnvMapping map[string]string

	// Follow indicates that logs should be streamed until cancelled.
	Follow bool

	// Since indicates from what time, relative to now, logs should be streamed.
	Since time.Duration

	// Since indicates until what time, relative to now, logs should be streamed.
	Until time.Duration

	// All indicates that we should return all containers regardless of their state.
	All bool

	// Limit restricts the total number of responses.
	Limit int

	// Filter filters the responses to the criteria set in it.
	Filter map[FilterKey][]string

	// Volumes is a list of volumes to attach to a container.
	Volumes []*cpb.Volume

	// VolumeDriverOptions is the driver options for the volume.
	VolumeDriverOptions proto.Message

	// VolumeLabels are optional labels that should be applied to the volume.
	VolumeLabels map[string]string
}

// WithTarget sets the target image name and tag option for this pull operation.
// Supported by: ContainerPush, ContainerPull
func WithTarget(image, tag string) Option {
	return func(p *options) {
		p.TargetName = image
		if tag == "" {
			tag = "latest"
		}
		p.TargetTag = tag
	}
}

// WithRegistryAuth sets the credentials to use for this pull operation.
// Supported by: ContainerPull
func WithRegistryAuth(creds *tpb.Credentials) Option {
	return func(p *options) {
		p.Credentials = creds
	}
}

// WithStream sets the stream to use to return results to the caller.
// Supported by: ContainerPull
func WithStream(client Stream) Option {
	return func(p *options) {
		p.StreamClient = client
	}
}

// Force sets the force operation field in the image options.
// Supported by: ContainerRemove, ContainerStop
func Force() Option {
	return func(p *options) {
		p.Force = true
	}
}

// WithInstanceName sets the name of the instance of a container.
// Supported by: ContainerStart
func WithInstanceName(instance string) Option {
	return func(p *options) {
		p.InstanceName = instance
	}
}

// WithPorts specifies the set of exposed ports for a container.
// Supported by: ContainerStart
func WithPorts(portMappings map[uint32]uint32) Option {
	return func(p *options) {
		p.PortMapping = portMappings
	}
}

// WithEnv specifies the set environment variables to set in the container.
// Supported by: ContainerStart
func WithEnv(envMapping map[string]string) Option {
	return func(p *options) {
		p.EnvMapping = envMapping
	}
}

// Follow specifies whether logs should be followed.
func Follow() Option {
	return func(p *options) {
		p.Follow = true
	}
}

// WithUntil specifies until when to collect logs.
func WithUntil(t time.Duration) Option {
	return func(p *options) {
		p.Until = t
	}
}

// WithSince specifies from when to collect logs.
func WithSince(t time.Duration) Option {
	return func(p *options) {
		p.Since = t
	}
}

// WithFilter provides the filter option.
// Supported by: ContainerList, VolumeList
func WithFilter(filter map[FilterKey][]string) Option {
	return func(p *options) {
		p.Filter = filter
	}
}

// WithVolumes sets the volumes to attach to a container.
// Supported by: ContainerStart
func WithVolumes(volumes []*cpb.Volume) Option {
	return func(p *options) {
		p.Volumes = volumes
	}
}

// WithVolumeDriverOpts provides the volume driver options.
// Supported by: CreateVolume
func WithVolumeDriverOpts(opts proto.Message) Option {
	return func(p *options) {
		p.VolumeDriverOptions = opts
	}
}

// WithVolumeLabels provides the volume labels.
// Supported by: CreateVolume
func WithVolumeLabels(labels map[string]string) Option {
	return func(p *options) {
		p.VolumeLabels = labels
	}
}

// ApplyOptions sets the passed options.
func ApplyOptions(opts ...Option) *options { // NOLINT
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	return options
}
