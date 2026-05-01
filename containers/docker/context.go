package docker

import (
	"context"
	"io"
)

type startPluginHookKey struct{}
type startPluginHookFunc func(ctx context.Context,
	pluginReader io.ReadCloser) (io.ReadCloser, error)

// NewContextWithStartPluginHook attaches a hook that runs during StartPlugin RPC.
// this hook handles the incoming tar file (with the embedded config.json)
// before the StartPlugin RPC uses that tar file to create+start a plugin.
func NewContextWithStartPluginHook(ctx context.Context,
	startPluginHook startPluginHookFunc) context.Context {
	return context.WithValue(ctx, startPluginHookKey{}, startPluginHook)
}

func startPluginHookFromContext(ctx context.Context) startPluginHookFunc {
	v, ok := ctx.Value(startPluginHookKey{}).(startPluginHookFunc)
	if !ok {
		return nil
	}
	return v
}
