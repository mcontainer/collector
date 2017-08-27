package docker

import (
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"context"
	"github.com/docker/docker/api/types"
	"log"
)

type IDockerClient interface {
	streamEvents(options types.EventsOptions) (<-chan events.Message, <-chan error)
	listNetworks(options types.NetworkListOptions) ([]types.NetworkResource, error)
	listContainers(options types.ContainerListOptions) ([]types.Container, error)
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

func (client *DockerClient) streamEvents(options types.EventsOptions) (<-chan events.Message, <-chan error) {
	ctx := context.Background()
	return client.api.Events(ctx, options)
}

func (client *DockerClient) listNetworks(options types.NetworkListOptions) ([]types.NetworkResource, error) {
	ctx := context.Background()
	return client.api.NetworkList(ctx, options)
}

func (client *DockerClient) listContainers(options types.ContainerListOptions) ([]types.Container, error) {
	ctx := context.Background()
	return client.api.ContainerList(ctx, options)
}
