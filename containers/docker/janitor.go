package docker

import (
	"context"
	"sync"
	"time"

	"github.com/moby/moby/v/v24/api/types/filters"
	"k8s.io/klog/v2"
)

var (
	cleaningInterval = 24 * time.Hour
)

// Vacuum cleans a docker container runtime.
type Vacuum struct {
	cli  docker
	quit chan struct{}
	wg   sync.WaitGroup
}

// NewJanitor creates a new docker janitor.
func NewJanitor(cli docker) *Vacuum {
	return &Vacuum{
		cli:  cli,
		quit: make(chan struct{}),
	}
}

// Start instructs the janitor to start working
func (j *Vacuum) Start(ctx context.Context) {
	klog.Info("janitor-starting")
	j.wg.Add(1)
	go j.vacuum(ctx)
	klog.Info("janitor-started")

}

// Stop instructs to janitor to take a rest.
func (j *Vacuum) Stop(ctx context.Context) {
	klog.Info("janitor-stopping")
	close(j.quit)
	j.wg.Wait()
	klog.Info("janitor-stopped")
}

// vacuum removes any dangling containers and images. Dangling containers are containers that
// have been stopped but not removed. Dangling images are intermediate images that were either
// used as part of a build or run that have no name, i.e. images with name '<none>'.
func (j *Vacuum) vacuum(ctx context.Context) {
	tick := time.NewTicker(cleaningInterval)
	defer tick.Stop()
	defer j.wg.Done()
	for {
		select {
		case <-j.quit:
			klog.Info("janitor was told to quit so it is")
			return
		case <-tick.C:
			cntReport, err := j.cli.ContainersPrune(ctx, filters.NewArgs())
			if err != nil {
				klog.Error("unable to vacuum containers %v", err)
			}

			imgReport, err := j.cli.ImagesPrune(ctx, filters.NewArgs())
			if err != nil {
				klog.Error("unable to vacuum images %v", err)
			}

			klog.Infof("Removed %d containers reclaiming %d bytes", len(cntReport.ContainersDeleted), cntReport.SpaceReclaimed)
			klog.Infof("Removed %d images reclaiming %d bytes", len(imgReport.ImagesDeleted), imgReport.SpaceReclaimed)
		}
	}
}
