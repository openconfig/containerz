// Package options provides options that can be used by container orchestration implementation.
package options

import (
	"time"

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

// ListStreamer is an entity capable of streaming container information.
type ListStreamer interface {
	Send(msg *cpb.ListResponse) error
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
)

// ImageOption takes an option and applies it to the set of options when the function is called.
type ImageOption func(*imageOptions)

// imageOptions contain the supported set of options for pull operations.
type imageOptions struct {

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
}

// WithTarget sets the target image name and tag option for this pull operation.
// Supported by: ContainerPush, ContainerPull
func WithTarget(image, tag string) ImageOption {
	return func(p *imageOptions) {
		p.TargetName = image
		if tag == "" {
			tag = "latest"
		}
		p.TargetTag = tag
	}
}

// WithRegistryAuth sets the credentials to use for this pull operation.
// Supported by: ContainerPull
func WithRegistryAuth(creds *tpb.Credentials) ImageOption {
	return func(p *imageOptions) {
		p.Credentials = creds
	}
}

// WithStream sets the stream to use to return results to the caller.
// Supported by: ContainerPull
func WithStream(client Stream) ImageOption {
	return func(p *imageOptions) {
		p.StreamClient = client
	}
}

// Force sets the force operation field in the image options.
// Supported by: ContainerRemove, ContainerStop
func Force() ImageOption {
	return func(p *imageOptions) {
		p.Force = true
	}
}

// WithInstanceName sets the name of the instance of a container.
// Supported by: ContainerStart
func WithInstanceName(instance string) ImageOption {
	return func(p *imageOptions) {
		p.InstanceName = instance
	}
}

// WithPorts specifies the set of exposed ports for a container.
// Supported by: ContainerStart
func WithPorts(portMappings map[uint32]uint32) ImageOption {
	return func(p *imageOptions) {
		p.PortMapping = portMappings
	}
}

// WithEnv specifies the set environment variables to set in the container.
// Supported by: ContainerStart
func WithEnv(envMapping map[string]string) ImageOption {
	return func(p *imageOptions) {
		p.EnvMapping = envMapping
	}
}

// Follow specifies whether logs should be followed.
func Follow() ImageOption {
	return func(p *imageOptions) {
		p.Follow = true
	}
}

// WithUntil specifies until when to collect logs.
func WithUntil(t time.Duration) ImageOption {
	return func(p *imageOptions) {
		p.Until = t
	}
}

// WithSince specifies from when to collect logs.
func WithSince(t time.Duration) ImageOption {
	return func(p *imageOptions) {
		p.Since = t
	}
}

// WithFilter provides the filter option.
// Supported by: ContainerList
func WithFilter(filter map[FilterKey][]string) ImageOption {
	return func(p *imageOptions) {
		p.Filter = filter
	}
}

// ApplyOptions sets the passed options.
func ApplyOptions(opts ...ImageOption) *imageOptions { // NOLINT
	options := &imageOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return options
}
