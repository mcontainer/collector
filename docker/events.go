package docker

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
	"time"
	"log"
)

type Docker struct {
	ctx     *context.Context
	cli     *client.Client
	filters *filters.Args
	Errors  <-chan error
	Data    chan types.Container
}

const (
	ACTION_CREATE = "create"
)

func NewDockerClient() *Docker {
	log.Println("Start creating docker client")
	c, _ := client.NewEnvClient()
	ctx := context.Background()
	f := filters.NewArgs()
	out := make(chan types.Container)
	outErr := make(chan error)
	d := Docker{
		ctx:     &ctx,
		cli:     c,
		filters: &f,
		Errors:  outErr,
		Data:    out,
	}
	return &d
}

func (docker *Docker) Listener() {
	docker.filters.Add("event", ACTION_CREATE)
	ev, err := docker.cli.Events(*docker.ctx, types.EventsOptions{Filters: *docker.filters })

	docker.Errors = err

	go func() {
		for {
			data := <-ev
			time.Sleep(500 * time.Millisecond)
			container := docker.filter(data.ID)
			docker.Data <- *container
		}
	}()
}

func (docker *Docker) filter(id string) *types.Container {
	list, err := docker.cli.ContainerList(*docker.ctx, types.ContainerListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, elm := range list {
		if elm.ID == id {
			return &elm
		}
	}
	return nil
}
