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
	"fmt"
	"math/big"
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

// ListImageStreamer is an entity capable of streaming container information.
type ListImageStreamer interface {
	Send(msg *cpb.ListImageResponse) error
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

	// Network is an option parameter that should be applied to a container upon startup. It
	// it represents the network to attach this container to. This could be 'host', 'bridged', or any
	// other network available in the runtime.
	Network string

	// Capabilities to be added/removed. Capabilities are first removed then added.
	Capabilities proto.Message

	// RestartPolicy to be assigned to the container.
	RestartPolicy proto.Message

	// RunAs provides a user (and potentially group) to run the container as.
	RunAs proto.Message

	// Labels is the set of labels or metadata to attach to the container.
	Labels map[string]string

	// IsPlugin indicates that this is only a plugin and therefore the tar file
	// should be saved only and not loaded into the container runtime.
	IsPlugin bool

	// CPU is the CPU limit for the container.
	CPU float64

	// SoftMemory is the soft memory limit for the container.
	SoftMemory int64

	// HardMemory is the hard memory limit for the container.
	HardMemory int64
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

// WithNetwork provides the network to attach this container to.
// Supported by: ContainerStart, ContainerUpdate
func WithNetwork(network string) Option {
	return func(p *options) {
		p.Network = network
	}
}

// WithCapabilities provides optional lists of added/removed container capabilities.
// Supported by: ContainerStart, ContainerUpdate
func WithCapabilities(opts proto.Message) Option {
	return func(p *options) {
		p.Capabilities = opts
	}
}

// WithRestartPolicy provides an optional restart policy for the container.
// Supported by: ContainerStart, ContainerUpdate
func WithRestartPolicy(opts proto.Message) Option {
	return func(p *options) {
		p.RestartPolicy = opts
	}
}

// WithRunAs provides an optional user (and potentially group) to run the container as.
// Supported by: ContainerStart, ContainerUpdate
func WithRunAs(opts proto.Message) Option {
	return func(p *options) {
		p.RunAs = opts
	}
}

// WithLabels provides an optional set of labels to attach to the container.
// Supported by: ContainerStart, ContainerUpdate
func WithLabels(labels map[string]string) Option {
	return func(p *options) {
		p.Labels = labels
	}
}

// WithCPUs provides the CPU limit for the container.
// Supported by: ContainerStart, ContainerUpdate
func WithCPUs(cpus float64) Option {
	return func(p *options) {
		p.CPU = cpus
	}
}

// WithSoftLimit provides the soft memory limit (in bytes) for the container.
// Supported by: ContainerStart, ContainerUpdate
func WithSoftLimit(mem int64) Option {
	return func(p *options) {
		p.SoftMemory = mem
	}
}

// WithHardLimit provides the hard memory limit (in bytes) for the container.
// Supported by: ContainerStart, ContainerUpdate
func WithHardLimit(mem int64) Option {
	return func(p *options) {
		p.HardMemory = mem
	}
}

// ParseCPUs takes a float returns an integer value of nano cpus
func ParseCPUs(value float64) (int64, error) {
	cpu := new(big.Rat).SetFloat64(value)

	nano := cpu.Mul(cpu, big.NewRat(1e9, 1))
	if !nano.IsInt() {
		return 0, fmt.Errorf("value is too precise")
	}
	return nano.Num().Int64(), nil
}

// ApplyOptions sets the passed options.
func ApplyOptions(opts ...Option) *options { // NOLINT
	options := &options{}
	for _, opt := range opts {
		opt(options)
	}

	return options
}
