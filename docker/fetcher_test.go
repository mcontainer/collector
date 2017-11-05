package docker

import (
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"testing"
)

type FakeClient struct {
	m mock.Mock
}

var (
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
		ID:         "123",
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
