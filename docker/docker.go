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
	Action    string
	Container types.Container
}

type Docker struct {
	Cli       *client.Client
	filters   *filters.Args
	IngressId string
	Errors    <-chan error
	Data      chan types.Container
	Stop      chan string
	NetworkID chan string
}

const (
	ACTION_CREATE = "create"
	ACTION_STOP   = "stop"
	ACTION_KILL   = "kill"
	ACTION_DIE    = "die"
)

func NewDockerClient() *Docker {
	log.Info("Start creating docker client")
	c, _ := client.NewEnvClient()
	f := filters.NewArgs()
	out := make(chan types.Container)
	networkCh := make(chan string)
	stop := make(chan string)
	outErr := make(chan error)
	return &Docker{
		Cli:       c,
		filters:   &f,
		Errors:    outErr,
		Data:      out,
		Stop:      stop,
		NetworkID: networkCh,
	}
}

func (docker *Docker) init() {
	docker.IngressId = docker.findIngressID()
}

func (docker *Docker) ListenSwarm() {
	ctx := context.Background()
	docker.init()
	docker.filters.Add("type", "container")
	docker.filters.Add("event", ACTION_CREATE)
	docker.filters.Add("event", ACTION_STOP)
	docker.filters.Add("event", ACTION_KILL)
	docker.filters.Add("event", ACTION_DIE)
	ev, err := docker.Cli.Events(ctx, types.EventsOptions{Filters: *docker.filters })
	docker.Errors = err

	for {
		data := <-ev
		if data.Action == ACTION_CREATE {
			log.WithField("Id", data.ID).Info("CREATE")
			services, err := docker.Cli.ServiceList(ctx, types.ServiceListOptions{})
			if err != nil {
				log.Fatal(err)
			}

			for _, service := range services {
				log.WithField("ServiceId", service.ID).Info()
				for _, ips := range service.Endpoint.VirtualIPs {
					if ips.NetworkID != docker.IngressId {
						log.WithField("Network ID", ips.NetworkID).Info("NETWORK")
						docker.NetworkID <- ips.NetworkID
					}
				}
			}

		}
	}

}

func (docker *Docker) findIngressID() string {
	ctx := context.Background()
	var id string;
	networks, _ := docker.Cli.NetworkList(ctx, types.NetworkListOptions{})
	for _, net := range networks {
		if net.Driver == "overlay" {
			if net.Name != "ingress" {
				log.WithField("Name", net.Name)
			} else {
				id = net.ID
			}
		}
	}
	log.WithField("ID", id).Info("Find Ingress network id")
	return id
}

func (docker *Docker) filter(ctx context.Context, id string) (*types.Container, error) {
	list, err := docker.Cli.ContainerList(ctx, types.ContainerListOptions{})
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
