package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"log"
)

type IDockerClient interface {
	StreamEvents(options types.EventsOptions) (<-chan events.Message, <-chan error)
	ListNetworks(options types.NetworkListOptions) ([]types.NetworkResource, error)
	ListContainers(options types.ContainerListOptions) ([]types.Container, error)
	InspectNetwork(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
}

type DockerClient struct {
	api *client.Client
}

func NewDockerClient() IDockerClient {
	c, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}
	return &DockerClient{api: c}
}

func (client *DockerClient) InspectNetwork(networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error) {
	ctx := context.Background()
	return client.api.NetworkInspect(ctx, networkID, options)
}

func (client *DockerClient) StreamEvents(options types.EventsOptions) (<-chan events.Message, <-chan error) {
	ctx := context.Background()
	return client.api.Events(ctx, options)
}

func (client *DockerClient) ListNetworks(options types.NetworkListOptions) ([]types.NetworkResource, error) {
	ctx := context.Background()
	return client.api.NetworkList(ctx, options)
}

func (client *DockerClient) ListContainers(options types.ContainerListOptions) ([]types.Container, error) {
	ctx := context.Background()
	return client.api.ContainerList(ctx, options)
}
