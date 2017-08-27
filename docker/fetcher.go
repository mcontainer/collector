package docker

import (
	"golang.org/x/net/context"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"errors"
)

type EventMessage struct {
	ContainerId string
	NetworkId   string
}

type Fetcher struct {
	cli       IDockerClient
	IngressId string
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

func NewFetcher(client IDockerClient) *Fetcher {
	log.Info("Start creating docker client")
	stop := make(chan string)
	return &Fetcher{
		cli:  client,
		Stop: stop,
	}
}

func (docker *Fetcher) initialize() (e error) {
	docker.IngressId, e = docker.findIngressID()
	return e
}

func (docker *Fetcher) Listen() (<-chan EventMessage, <-chan error) {
	ctx := context.Background()
	outChan := make(chan EventMessage)
	if e := docker.initialize();e != nil {
		log.Fatal(e)
	}

	f := filters.NewArgs()
	f.Add("type", "container")
	f.Add("event", ACTION_CREATE)
	f.Add("event", ACTION_STOP)
	f.Add("event", ACTION_KILL)
	f.Add("event", ACTION_DIE)
	f.Add("event", ACTION_START)
	ev, errChan := docker.cli.streamEvents(types.EventsOptions{Filters: f })

	go func() {
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
						outChan <- EventMessage{
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
	}()

	return outChan, errChan

}

func (docker *Fetcher) findIngressID() (string, error) {
	f := filters.NewArgs()
	f.Add("name", INGRESS)
	networks, _ := docker.cli.listNetworks(types.NetworkListOptions{Filters: f})
	if len(networks) == 0 {
		return "", errors.New("Ingress network ID not found")
	}
	log.WithField("ID", networks[0].ID).Info("Find Ingress network id")
	return networks[0].ID, nil
}

func (docker *Fetcher) filter(ctx context.Context, id string) (*types.Container, error) {
	f := filters.NewArgs()
	f.Add("id", id)
	list, err := docker.cli.listContainers(types.ContainerListOptions{Filters: f})
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
