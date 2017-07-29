package docker

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types"
	"time"
	log "github.com/sirupsen/logrus"
	"errors"
	"github.com/Workiva/go-datastructures/bitarray"
	"fmt"
)

type EventMessage struct {
	Action    string
	Container types.Container
}

type Docker struct {
	cli           *client.Client
	filters       *filters.Args
	Errors        <-chan error
	Data          chan types.Container
	Stop          chan string
	InMemoryPorts *map[string]bitarray.BitArray
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
	stop := make(chan string)
	outErr := make(chan error)
	mapPorts := make(map[string]bitarray.BitArray)
	return &Docker{
		cli:           c,
		filters:       &f,
		Errors:        outErr,
		Data:          out,
		Stop:          stop,
		InMemoryPorts: &mapPorts,
	}
}

func (docker *Docker) Listen() {
	ctx := context.Background()
	docker.filters.Add("event", ACTION_CREATE)
	docker.filters.Add("event", ACTION_STOP)
	docker.filters.Add("event", ACTION_KILL)
	docker.filters.Add("event", ACTION_DIE)
	ev, err := docker.cli.Events(ctx, types.EventsOptions{Filters: *docker.filters })

	docker.Errors = err

	for {
		data := <-ev
		if data.Action == ACTION_CREATE {
			time.Sleep(500 * time.Millisecond)
			container, e := docker.filter(ctx, data.ID)
			if e != nil {
				log.Fatal(e)
			} else {
				docker.Data <- *container
			}
		} else {
			log.WithField("message", data).Warning("Container died / stopped / killed")
			log.WithField("remaining ports", (*docker.InMemoryPorts)[data.ID].ToNums()).Info("Ports")
			docker.Stop <- data.ID
			log.WithField("message", data).Warning("Container stopped")
		}
	}
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

func (docker *Docker) ToBitPorts(container types.Container) bitarray.BitArray {
	b := bitarray.NewSparseBitArray()
	for _, p := range container.Ports {
		b.SetBit(uint64(p.PrivatePort))
	}
	fmt.Println(container.ID)
	fmt.Println(container.Ports)
	(*docker.InMemoryPorts)[container.ID] = b
	log.WithField("InMemoryPorts", (*docker.InMemoryPorts)[container.ID].ToNums()).Info("Set InMemoryPorts")
	return b
}

func UpdatePort(bPortsContainer bitarray.BitArray, bActualPorts bitarray.BitArray) (bitarray.BitArray, bool) {
	return bActualPorts.Or(bPortsContainer), !bActualPorts.Equals(bPortsContainer)
}

func (docker *Docker) RemovePorts(id string, b bitarray.BitArray) bitarray.BitArray {
	a := (*docker.InMemoryPorts)[id]

	ports := a.ToNums()
	for _, v := range ports {
		b.ClearBit(v)
	}

	return b

}
