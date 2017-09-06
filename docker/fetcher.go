package docker

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
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
	ACTION_CREATE  = "create"
	ACTION_CONNECT = "connect"
	ACTION_STOP    = "stop"
	ACTION_KILL    = "kill"
	ACTION_DIE     = "die"
	ACTION_START   = "start"
	INGRESS        = "ingress"
)

func NewFetcher(client IDockerClient) *Fetcher {
	log.Info("Fetcher:: Start creating fetcher")
	stop := make(chan string)
	return &Fetcher{
		cli:  client,
		Stop: stop,
	}
}

func (fetcher *Fetcher) ListenNetwork() (<-chan EventMessage, <-chan error) {
	networkChan := make(chan EventMessage)
	f := filters.NewArgs()
	f.Add("type", "network")
	f.Add("event", ACTION_CONNECT)
	events, err := fetcher.cli.streamEvents(types.EventsOptions{Filters: f})
	go func() {
		for {
			data := <-events
			switch data.Action {
			case ACTION_CONNECT:
				log.WithFields(log.Fields{
					"ID":        data.Actor.ID,
					"container": data.Actor.Attributes["container"],
				}).Info("Fetcher::Network -- CONNECTION")
				networkChan <- EventMessage{
					ContainerId: data.Actor.Attributes["container"],
					NetworkId:   data.Actor.ID,
				}
			}
		}
	}()
	return networkChan, err
}

// TODO: refactor (not used at this time)
func (fetcher *Fetcher) Listen() (<-chan EventMessage, <-chan error) {
	ctx := context.Background()
	outChan := make(chan EventMessage)
	f := filters.NewArgs()
	f.Add("type", "container")
	f.Add("event", ACTION_CREATE)
	f.Add("event", ACTION_STOP)
	f.Add("event", ACTION_KILL)
	f.Add("event", ACTION_DIE)
	f.Add("event", ACTION_START)
	ev, errChan := fetcher.cli.streamEvents(types.EventsOptions{Filters: f})
	go func() {
		for {
			data := <-ev
			switch data.Action {
			case ACTION_START:
				log.WithField("Id", data.ID).Info("START")
				container, e := fetcher.FilterContainer(ctx, data.ID)
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

func (fetcher *Fetcher) FindOverlayNetworks() ([]types.NetworkResource, error) {
	f := filters.NewArgs()
	f.Add("driver", "overlay")
	networks, err := fetcher.cli.listNetworks(types.NetworkListOptions{Filters: f})
	if err != nil {
		return nil, err
	}
	if len(networks) == 0 {
		return nil, errors.New("Not found networks")
	}
	log.WithField("size", len(networks)).Info("Fetcher:: Found overlay networks")
	return networks, nil
}

func (fetcher *Fetcher) FilterContainer(ctx context.Context, id string) (*types.Container, error) {
	f := filters.NewArgs()
	f.Add("id", id)
	containers, err := fetcher.cli.listContainers(types.ContainerListOptions{Filters: f})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, errors.New("No containers")
	}
	return &containers[0], nil
}
