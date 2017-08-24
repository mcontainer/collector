package docker

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"errors"
)

type EventMessage struct {
	ContainerId string
	NetworkId   string
}

type Docker struct {
	Cli       *client.Client
	IngressId string
	Errors    <-chan error
	Data      chan EventMessage
	Stop      chan string
}

const (
	ACTION_CREATE = "create"
	ACTION_STOP   = "stop"
	ACTION_KILL   = "kill"
	ACTION_DIE    = "die"
	ACTION_START  = "start"
	INGRESS       = "ingress"
)

func NewDockerClient() *Docker {
	log.Info("Start creating docker client")
	c, _ := client.NewEnvClient()
	out := make(chan EventMessage)
	stop := make(chan string)
	outErr := make(chan error)
	return &Docker{
		Cli:    c,
		Errors: outErr,
		Data:   out,
		Stop:   stop,
	}
}

func (docker *Docker) init() {
	docker.IngressId = docker.findIngressID()
}

func (docker *Docker) Listen() {
	ctx := context.Background()
	docker.init()
	f := filters.NewArgs()
	f.Add("type", "container")
	f.Add("event", ACTION_CREATE)
	f.Add("event", ACTION_STOP)
	f.Add("event", ACTION_KILL)
	f.Add("event", ACTION_DIE)
	f.Add("event", ACTION_START)
	ev, err := docker.Cli.Events(ctx, types.EventsOptions{Filters: f })
	docker.Errors = err

	for {
		data := <-ev
		switch data.Action {
		case ACTION_START:
			log.WithField("Id", data.ID).Info("START")
			container, e := docker.filter(ctx, data.ID)
			if e != nil {
				log.Fatal(e)
			}
			for k, v := range container.NetworkSettings.Networks {
				if k != INGRESS {
					docker.Data <- EventMessage{
						ContainerId: container.ID,
						NetworkId:   v.NetworkID,
					}
				}
			}
		case ACTION_CREATE:
			log.WithField("Id", data.ID).Info("CREATE")
			//TODO: make grpc callS
		}
	}

}

func (docker *Docker) findIngressID() string {
	ctx := context.Background()
	f := filters.NewArgs()
	f.Add("name", INGRESS)
	networks, _ := docker.Cli.NetworkList(ctx, types.NetworkListOptions{
		Filters: f,
	})
	if len(networks) == 0 {
		log.Fatal("Ingress Network not found")
	}
	log.WithField("ID", networks[0].ID).Info("Find Ingress network id")
	return networks[0].ID
}

func (docker *Docker) filter(ctx context.Context, id string) (*types.Container, error) {
	f := filters.NewArgs()
	f.Add("id", id)
	list, err := docker.Cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: f,
	})
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
