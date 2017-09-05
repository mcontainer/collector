package docker

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

type FakeClient struct{}

var (
	fakeClient *FakeClient
	fetcher    *Fetcher
	networks   []types.NetworkResource
	containers []types.Container
	event      []events.Message
)

func (c *FakeClient) listNetworks(options types.NetworkListOptions) ([]types.NetworkResource, error) {
	for _, network := range networks {
		if network.Name == options.Filters.Get("name")[0] {
			return []types.NetworkResource{network}, nil
		}
	}
	return []types.NetworkResource{}, errors.New("not found")
}

func (c *FakeClient) listContainers(options types.ContainerListOptions) ([]types.Container, error) {
	for _, container := range containers {
		if container.ID == options.Filters.Get("id")[0] {
			return []types.Container{container}, nil
		}
	}
	return nil, errors.New("Not found")
}

func (c *FakeClient) streamEvents(options types.EventsOptions) (<-chan events.Message, <-chan error) {
	messages := make(chan events.Message)
	errs := make(chan error, 1)
	started := make(chan struct{})
	go func() {
		defer close(errs)
		close(started)
		for _, event := range event {
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
	fetcher = NewFetcher(&FakeClient{})
}

func TestNewFetcher(t *testing.T) {
	client := NewDockerClient()
	fetcher := NewFetcher(client)
	assert.NotNil(t, fetcher)
}

func TestFilterSucess(t *testing.T) {
	ctx := context.Background()
	containers = []types.Container{
		{
			ID: "123a",
		},
		{
			ID: "456b",
		},
	}
	container, error := fetcher.filter(ctx, "456b")
	assert.Empty(t, error)
	assert.NotEmpty(t, container)
	assert.Equal(t, containers[1].ID, container.ID)
}

func TestFilterFailure(t *testing.T) {
	ctx := context.Background()
	containers = []types.Container{
		{
			ID: "123a",
		},
		{
			ID: "456b",
		},
	}
	container, error := fetcher.filter(ctx, "123")
	assert.Empty(t, container)
	assert.NotEmpty(t, error)
}

func TestListen(t *testing.T) {

	endpoints := make(map[string]*network.EndpointSettings)
	endpoints["toto"] = &network.EndpointSettings{NetworkID: "154a"}
	endpoints["titi"] = &network.EndpointSettings{NetworkID: "456u"}

	event = []events.Message{
		{
			Action: "create",
			ID:     "123a",
		},
		{
			Action: "start",
			ID:     "456b",
		},
	}

	containers = []types.Container{
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
	networks = []types.NetworkResource{
		{
			Name: "toto",
			ID:   "12345",
		},
		{
			Name: "ingress",
			ID:   "123456789",
		},
	}
	out, _ := fetcher.Listen()

	data := <-out
	assert.NotNil(t, data)
	assert.Equal(t, containers[1].ID, data.ContainerId)
	assert.Equal(t, containers[0].NetworkSettings.Networks["toto"].NetworkID, data.NetworkId)

	data = <-out
	assert.NotNil(t, data)
	assert.Equal(t, containers[1].ID, data.ContainerId)
	assert.Equal(t, containers[0].NetworkSettings.Networks["titi"].NetworkID, data.NetworkId)

}
