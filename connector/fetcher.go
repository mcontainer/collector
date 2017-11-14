package connector

import (
	"docker-visualizer/collector/connector/docker"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type EventMessage struct {
	Action      string
	ContainerId string
	NetworkId   string
	NetworkName string
}

type Fetcher struct {
	cli       docker.IDockerClient
	IngressId string
	Stop      chan string
}

const (
	ACTION_CONNECT    = "connect"
	ACTION_DISCONNECT = "disconnect"
	INGRESS           = "ingress"
)

func NewFetcher(client docker.IDockerClient) *Fetcher {
	log.Info("Creating fetcher")
	stop := make(chan string)
	return &Fetcher{
		cli:  client,
		Stop: stop,
	}
}

func (fetcher *Fetcher) DockerFromNetwork(id string) ([]*types.Container, error) {
	ctx := context.Background()
	var containers []*types.Container
	network, err := fetcher.cli.InspectNetwork(id, types.NetworkInspectOptions{})
	if err != nil {
		return nil, err
	}
	for k := range network.Containers {
		container, err := fetcher.FilterContainer(ctx, k)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}
	return containers, nil
}

func (fetcher *Fetcher) ListenNetwork() (<-chan EventMessage, <-chan error) {
	networkChan := make(chan EventMessage)
	f := filters.NewArgs()
	f.Add("type", "network")
	f.Add("event", ACTION_CONNECT)
	f.Add("event", ACTION_DISCONNECT)
	events, err := fetcher.cli.StreamEvents(types.EventsOptions{Filters: f})
	go func() {
		for {
			data := <-events
			switch data.Action {
			case ACTION_CONNECT:
				if data.Actor.ID != fetcher.IngressId && data.Actor.Attributes["name"] != INGRESS {
					log.WithFields(log.Fields{
						"ID":        data.Actor.ID,
						"container": data.Actor.Attributes["container"],
					}).Info("Fetcher::Network -- CONNECTION")
					networkChan <- EventMessage{
						Action:      ACTION_CONNECT,
						ContainerId: data.Actor.Attributes["container"],
						NetworkId:   data.Actor.ID,
						NetworkName: data.Actor.Attributes["name"],
					}
				}
			case ACTION_DISCONNECT:
				log.WithFields(log.Fields{
					"ID":        data.Actor.ID,
					"container": data.Actor.Attributes["container"],
				}).Info("Fetcher::Network -- DISCONNECTION")
				networkChan <- EventMessage{
					Action:      ACTION_DISCONNECT,
					ContainerId: data.Actor.Attributes["container"],
					NetworkId:   data.Actor.ID,
				}
			}
		}
	}()
	return networkChan, err
}

func (fetcher *Fetcher) FindOverlayNetworks() ([]types.NetworkResource, error) {
	f := filters.NewArgs()
	f.Add("driver", "overlay")
	networks, err := fetcher.cli.ListNetworks(types.NetworkListOptions{Filters: f})
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
	containers, err := fetcher.cli.ListContainers(types.ContainerListOptions{Filters: f})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, errors.New("No containers")
	}
	return &containers[0], nil
}
