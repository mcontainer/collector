package docker

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/docker/docker/api/types"
	"context"
	"github.com/docker/docker/api/types/filters"
	"errors"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/network"
)

type FakeClient struct{}

type FakeDocker struct {
	Cli        *FakeClient
	IngressId  string
	Errors     <-chan error
	Data       chan EventMessage
	Stop       chan string
	networks   []types.NetworkResource
	containers []types.Container
	events     []events.Message
}

var (
	fakeClient *FakeClient
	fakeDocker *FakeDocker
)

func (c *FakeClient) NetworkList(ctx context.Context, options types.NetworkListOptions) ([]types.NetworkResource, error) {
	for _, network := range fakeDocker.networks {
		if network.Name == options.Filters.Get("name")[0] {
			return []types.NetworkResource{network}, nil
		}
	}
	return []types.NetworkResource{}, errors.New("not found")
}

func (c *FakeClient) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	for _, container := range fakeDocker.containers {
		if container.ID == options.Filters.Get("id")[0] {
			return []types.Container{container}, nil
		}
	}
	return nil, errors.New("Not found")
}

func (c *FakeClient) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	messages := make(chan events.Message)
	errs := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		defer close(errs)
		close(started)
		for _, event := range fakeDocker.events {
			switch event.Action {
			case "error":
				errs <- errors.New("chan error")
			default:
				for _, filter := range options.Filters.Get("event") {
					if filter == event.Action {
						messages <- event
					}
				}
			}
		}
	}()
	<-started
	return messages, errs
}

func init() {
	out := make(chan EventMessage)
	fakeClient = &FakeClient{}
	fakeDocker = &FakeDocker{
		Data: out,
	}
	fakeDocker.Cli = fakeClient
}

func (c *FakeDocker) initialize() {
	c.IngressId, _ = c.findIngressID()
}

func (c *FakeDocker) findIngressID() (string, error) {
	ctx := context.Background()
	f := filters.NewArgs()
	f.Add("name", "ingress")
	networks, _ := c.Cli.NetworkList(ctx, types.NetworkListOptions{
		Filters: f,
	})
	if len(networks) == 0 {
		return "", errors.New("not found")
	}
	return networks[0].ID, nil
}

func (c *FakeDocker) filter(ctx context.Context, id string) (*types.Container, error) {
	f := filters.NewArgs()
	f.Add("id", id)
	list, err := c.Cli.ContainerList(ctx, types.ContainerListOptions{
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

func (c *FakeDocker) Listen() {
	ctx := context.Background()
	c.initialize()
	f := filters.NewArgs()
	f.Add("type", "container")
	f.Add("event", ACTION_CREATE)
	f.Add("event", ACTION_STOP)
	f.Add("event", ACTION_KILL)
	f.Add("event", ACTION_DIE)
	f.Add("event", ACTION_START)
	ev, err := c.Cli.Events(ctx, types.EventsOptions{Filters: f })
	fakeDocker.Errors = err

	for {
		data := <-ev
		switch data.Action {
		case ACTION_START:
			container, e := fakeDocker.filter(ctx, data.ID)
			if e != nil {
				return
			}
			for k, v := range container.NetworkSettings.Networks {
				if k != INGRESS {
					fakeDocker.Data <- EventMessage{
						ContainerId: container.ID,
						NetworkId:   v.NetworkID,
					}
				}
			}
		case ACTION_CREATE:
			container, e := fakeDocker.filter(ctx, data.ID)
			if e != nil {
				return
			}
			for k, v := range container.NetworkSettings.Networks {
				if k != INGRESS {
					fakeDocker.Data <- EventMessage{
						ContainerId: container.ID,
						NetworkId:   v.NetworkID,
					}
				}
			}
			//TODO: make grpc callS
		}
	}
}

func TestNewDockerClient(t *testing.T) {
	client := NewDockerClient()
	assert.NotNil(t, client)
}

func TestFindIngressId(t *testing.T) {
	fakeDocker.networks = []types.NetworkResource{
		{
			Name: "toto",
			ID:   "12345",
		},
		{
			Name: "ingress",
			ID:   "123456789",
		},
	}

	id, e := fakeDocker.findIngressID()

	assert.Nil(t, e)
	assert.NotEmpty(t, id)
	assert.Equal(t, fakeDocker.networks[1].ID, id)

}

func TestNotFindIngressId(t *testing.T) {
	fakeDocker.networks = []types.NetworkResource{
		{
			Name: "toto",
			ID:   "12345",
		},
	}
	id, e := fakeDocker.findIngressID()
	assert.Empty(t, id)
	assert.NotNil(t, e)
}

func TestInitializeSuccess(t *testing.T) {
	fakeDocker.networks = []types.NetworkResource{
		{
			Name: "toto",
			ID:   "12345",
		},
		{
			Name: "ingress",
			ID:   "123456789",
		},
	}
	fakeDocker.initialize()
	assert.NotEmpty(t, fakeDocker.IngressId)
	assert.Equal(t, fakeDocker.networks[1].ID, fakeDocker.IngressId)
}

func TestInitializeFailure(t *testing.T) {
	fakeDocker.networks = []types.NetworkResource{
		{
			Name: "toto",
			ID:   "12345",
		},
	}
	fakeDocker.initialize()
	assert.Empty(t, fakeDocker.IngressId)
	assert.Equal(t, "", fakeDocker.IngressId)
}

func TestFilterSucess(t *testing.T) {
	ctx := context.Background()
	fakeDocker.containers = []types.Container{
		{
			ID: "123a",
		},
		{
			ID: "456b",
		},
	}
	container, error := fakeDocker.filter(ctx, "456b")
	assert.Empty(t, error)
	assert.NotEmpty(t, container)
	assert.Equal(t, fakeDocker.containers[1].ID, container.ID)
}

func TestFilterFailure(t *testing.T) {
	ctx := context.Background()
	fakeDocker.containers = []types.Container{
		{
			ID: "123a",
		},
		{
			ID: "456b",
		},
	}
	container, error := fakeDocker.filter(ctx, "123")
	assert.Empty(t, container)
	assert.NotEmpty(t, error)
	assert.Equal(t, "Not found", error.Error())
}

func TestListen(t *testing.T) {

	endpoints := make(map[string]*network.EndpointSettings)
	endpoints["toto"] = &network.EndpointSettings{NetworkID: "154a"}
	endpoints["titi"] = &network.EndpointSettings{NetworkID: "456u"}

	fakeDocker.events = []events.Message{
		{
			Action: "create",
			ID:     "123a",
		},
		{
			Action: "start",
			ID:     "456b",
		},
	}

	fakeDocker.containers = []types.Container{
		{
			ID: "123a",
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: endpoints,
			},
		},
		{
			ID: "456b",
			NetworkSettings: &types.SummaryNetworkSettings{
				Networks: endpoints,
			},
		},
	}
	fakeDocker.networks = []types.NetworkResource{
		{
			Name: "toto",
			ID:   "12345",
		},
		{
			Name: "ingress",
			ID:   "123456789",
		},
	}
	go fakeDocker.Listen()

	data := <-fakeDocker.Data
	assert.NotNil(t, data)
	assert.Equal(t, fakeDocker.containers[0].ID, data.ContainerId)
	assert.Equal(t, fakeDocker.containers[0].NetworkSettings.Networks["toto"].NetworkID, data.NetworkId)

	data = <-fakeDocker.Data
	assert.NotNil(t, data)
	assert.Equal(t, fakeDocker.containers[0].ID, data.ContainerId)
	assert.Equal(t, fakeDocker.containers[0].NetworkSettings.Networks["titi"].NetworkID, data.NetworkId)

	data = <-fakeDocker.Data
	assert.NotNil(t, data)
	assert.Equal(t, fakeDocker.containers[1].ID, data.ContainerId)
	assert.Equal(t, fakeDocker.containers[1].NetworkSettings.Networks["toto"].NetworkID, data.NetworkId)

	data = <-fakeDocker.Data
	assert.NotNil(t, data)
	assert.Equal(t, fakeDocker.containers[1].ID, data.ContainerId)
	assert.Equal(t, fakeDocker.containers[1].NetworkSettings.Networks["titi"].NetworkID, data.NetworkId)

}
