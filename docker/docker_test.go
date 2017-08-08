package docker

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

func TestNewDockerClient(t *testing.T) {
	client := NewDockerClient()
	assert.NotNil(t, client)
}

func TestDocker_Listen(t *testing.T) {
	ctx := context.Background()
	docker := NewDockerClient()

	docker.Listen()

	_, err := docker.Cli.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	resp, err := docker.Cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"sleep", "2"},
	}, nil, nil, "")

	if err != nil {
		panic(err)
	}

	if err := docker.Cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	found := <-docker.Data

	assert.Equal(t, resp.ID, found.ID)

}
