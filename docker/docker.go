package docker

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
	"time"
	log "github.com/sirupsen/logrus"
	"errors"
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
	log.Info("Start creating docker client")
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

func (docker *Docker) Listen() {
	ctx := context.Background()
	docker.filters.Add("event", ACTION_CREATE)
	ev, err := docker.cli.Events(ctx, types.EventsOptions{Filters: *docker.filters })

	docker.Errors = err

	go func() {
		for {
			data := <-ev
			time.Sleep(500 * time.Millisecond)
			container, e := docker.filter(ctx, data.ID)
			if e != nil {
				log.Fatal(e)
			} else {
				docker.Data <- *container
			}
		}
	}()
}

func (docker *Docker) filter(ctx context.Context, id string) (*types.Container, error) {
	list, err := docker.cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	for _, elm := range list {
		if elm.ID == id {
			return &elm, nil
		}
	}
	return nil, errors.New("Container " + id + " not found")
}
