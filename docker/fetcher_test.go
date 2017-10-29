package docker

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
	"github.com/stretchr/testify/mock"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

type FakeClient struct {
	m mock.Mock
}

var (
	networks   []types.NetworkResource
	containers []types.Container
	event      []events.Message
)

func (c *FakeClient) listNetworks(options types.NetworkListOptions) ([]types.NetworkResource, error) {
	args := c.m.Called(options)
	return args.Get(0).([]types.NetworkResource), args.Error(1)
}

func (c *FakeClient) inspectNetwork(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error) {
	args := c.m.Called(networkID, options)
	return args.Get(0).(types.NetworkResource), args.Error(1)
}

func (c *FakeClient) listContainers(options types.ContainerListOptions) ([]types.Container, error) {
	args := c.m.Called(options)
	return args.Get(0).([]types.Container), args.Error(1)
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

func TestNewFetcher(t *testing.T) {
	client := NewDockerClient()
	fetcher := NewFetcher(client)
	assert.NotNil(t, fetcher)
}

func TestFilterSucess(t *testing.T) {
	ctx := context.Background()
	containers = []types.Container{{ID: "456b"}}
	client := FakeClient{}
	f := filters.NewArgs()
	f.Add("id", "456b")
	client.m.On("listContainers", types.ContainerListOptions{Filters: f}).Return(containers, nil).Once()
	fetcher := NewFetcher(&client)
	container, e := fetcher.FilterContainer(ctx, "456b")
	assert.Empty(t, e)
	assert.NotEmpty(t, container)
	assert.Equal(t, containers[0].ID, container.ID)
}

func TestFilterFailure(t *testing.T) {
	ctx := context.Background()
	client := FakeClient{}
	f := filters.NewArgs()
	f.Add("id", "123")
	client.m.On("listContainers", types.ContainerListOptions{Filters: f}).Return([]types.Container{}, errors.New("not found"))
	fetcher := NewFetcher(&client)
	container, e := fetcher.FilterContainer(ctx, "123")
	assert.Empty(t, container)
	assert.NotEmpty(t, e)
}

func TestFindOverlaySuccess(t *testing.T) {
	mockClient := FakeClient{}
	f := filters.NewArgs()
	f.Add("driver", "overlay")
	mockResults := []types.NetworkResource{
		{
			ID: "123",
		},
	}
	mockClient.m.On("listNetworks", types.NetworkListOptions{Filters: f}).Return(mockResults, nil)
	fetcher := NewFetcher(&mockClient)
	networks, e := fetcher.FindOverlayNetworks()
	assert.Empty(t, e)
	assert.NotNil(t, networks)
	assert.Equal(t, mockResults[0], networks[0])
}

func TestFindOverlay0Size(t *testing.T) {
	mockClient := FakeClient{}
	f := filters.NewArgs()
	f.Add("driver", "overlay")
	mockClient.m.On("listNetworks", types.NetworkListOptions{Filters: f}).Return([]types.NetworkResource{}, nil)
	fetcher := NewFetcher(&mockClient)
	networks, e := fetcher.FindOverlayNetworks()
	assert.NotEmpty(t, e)
	assert.Len(t, networks, 0)
	assert.Equal(t, "Not found networks", e.Error())
}

func TestFindOverlayFailure(t *testing.T) {
	mockClient := FakeClient{}
	f := filters.NewArgs()
	f.Add("driver", "overlay")
	mockClient.m.On("listNetworks", types.NetworkListOptions{Filters: f}).Return([]types.NetworkResource{}, errors.New("custom error"))
	fetcher := NewFetcher(&mockClient)
	networks, e := fetcher.FindOverlayNetworks()
	assert.NotEmpty(t, e)
	assert.Len(t, networks, 0)
	assert.Equal(t, "custom error", e.Error())
}

func TestDockerFromNetworkSucess(t *testing.T) {
	mockContainers := map[string]types.EndpointResource{}
	mockContainers["test"] = types.EndpointResource{}
	mockContainers["titi"] = types.EndpointResource{}
	mockResults := types.NetworkResource{
			ID: "123",
			Containers: mockContainers,
	}

	mockClient := FakeClient{}
	mockClient.m.On("inspectNetwork", "123", types.NetworkInspectOptions{}).Return(mockResults, nil)
	f := filters.NewArgs()
	f.Add("id", "test")
	mockClient.m.On("listContainers", types.ContainerListOptions{Filters: f}).Return([]types.Container{{ID: "test"}}, nil)
	f2 := filters.NewArgs()
	f2.Add("id", "titi")
	mockClient.m.On("listContainers", types.ContainerListOptions{Filters: f2}).Return([]types.Container{{ID: "titi"}}, nil)

	fetcher := NewFetcher(&mockClient)

	containers, e := fetcher.DockerFromNetwork("123")
	assert.Empty(t, e)
	assert.NotEmpty(t, containers)
	assert.Len(t, containers, 2)
	assert.Equal(t, "test", containers[0].ID)
	assert.Equal(t, "titi", containers[1].ID)
}

func TestListen(t *testing.T) {

	mockClient := FakeClient{}

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

	f := filters.NewArgs()
	f.Add("id", "123a")
	f2 := filters.NewArgs()
	f2.Add("id", "456b")
	mockClient.m.On("listContainers", types.ContainerListOptions{Filters: f}).Return([]types.Container{containers[0]}, nil)
	mockClient.m.On("listContainers", types.ContainerListOptions{Filters: f2}).Return([]types.Container{containers[1]}, nil)
	fetcher := NewFetcher(&mockClient)

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
