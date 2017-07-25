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
	f := filters.NewArgs()
	out := make(chan types.Container)
	outErr := make(chan error)
	return &Docker{
		cli:     c,
		filters: &f,
		Errors:  outErr,
		Data:    out,
	}
}

func (docker *Docker) Listener() {
	ctx := context.Background()
	docker.filters.Add("event", ACTION_CREATE)
	ev, err := docker.cli.Events(ctx, types.EventsOptions{Filters: *docker.filters })

	docker.Errors = err

	go func() {
		for {
			data := <-ev
			time.Sleep(500 * time.Millisecond)
			container := docker.filter(ctx, data.ID)
			docker.Data <- *container
		}
	}()
}

func (docker *Docker) filter(ctx context.Context, id string) *types.Container {
	list, err := docker.cli.ContainerList(ctx, types.ContainerListOptions{})
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
