package event

import (
	pb "docker-visualizer/proto/containers"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"testing"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) AddNode(ctx context.Context, in *pb.ContainerInfo, opts ...grpc.CallOption) (*pb.Response, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.Response), args.Error(1)
}

func (m *Mock) RemoveNode(ctx context.Context, in *pb.ContainerID, opts ...grpc.CallOption) (*pb.Response, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.Response), args.Error(1)
}

func (m *Mock) StreamContainerEvents(ctx context.Context, opts ...grpc.CallOption) (pb.ContainerService_StreamContainerEventsClient, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(pb.ContainerService_StreamContainerEventsClient), args.Error(1)
}

func TestNewEventBroker(t *testing.T) {
	mockObject := new(Mock)
	broker := NewEventBroker(mockObject)
	assert.NotNil(t, broker)
	assert.NotNil(t, broker.Stream)
	assert.NotNil(t, broker.grpc)
}

func TestEventBroker_SendNode(t *testing.T) {
	mockGrpc := new(Mock)
	broker := NewEventBroker(mockGrpc)
	mockGrpc.On("AddNode", mock.Anything, &pb.ContainerInfo{
		Id:      "123",
		Name:    "toto",
		Service: "server",
		Ip:      "10.0.0.3",
		Network: "test",
		Stack:   "microservice",
		Host:    "host",
	}, mock.Anything).Return(&pb.Response{Success: true}, nil)

	settings := make(map[string]*network.EndpointSettings)
	settings["test"] = &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: "10.0.0.3",
		},
	}

	e := broker.SendNode(&types.Container{
		ID:    "123",
		Names: []string{"toto"},
		NetworkSettings: &types.SummaryNetworkSettings{
			Networks: settings,
		},
		Labels: map[string]string{"com.docker.swarm.service.name": "server"},
	}, "test", "host")

	mockGrpc.AssertNumberOfCalls(t, "AddNode", 1)

	assert.Nil(t, e)

}

func TestEventBroker_SendNodeFailureConnection(t *testing.T) {
	mockGrpc := new(Mock)
	broker := NewEventBroker(mockGrpc)
	mockGrpc.On("AddNode", mock.Anything, &pb.ContainerInfo{
		Id:      "123",
		Name:    "toto",
		Service: "server",
		Ip:      "10.0.0.3",
		Network: "test",
		Stack:   "microservice",
		Host:    "host",
	}, mock.Anything).Return(&pb.Response{}, errors.New("error"))

	settings := make(map[string]*network.EndpointSettings)
	settings["test"] = &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: "10.0.0.3",
		},
	}

	e := broker.SendNode(&types.Container{
		ID:    "123",
		Names: []string{"toto"},
		NetworkSettings: &types.SummaryNetworkSettings{
			Networks: settings,
		},
		Labels: map[string]string{"com.docker.swarm.service.name": "server"},
	}, "test", "host")

	mockGrpc.AssertNumberOfCalls(t, "AddNode", 1)

	assert.NotNil(t, e)

}

func TestEventBroker_RemoveNode(t *testing.T) {
	m := new(Mock)
	b := NewEventBroker(m)
	m.On("RemoveNode", mock.Anything, &pb.ContainerID{Id: "123456"}, mock.Anything).
		Return(&pb.Response{Success: true}, nil)
	e := b.RemoveNode("123456")
	m.AssertNumberOfCalls(t, "RemoveNode", 1)
	assert.Nil(t, e)
}

func TestEventBroker_RemoveNodeFailureServerConnection(t *testing.T) {
	m := new(Mock)
	b := NewEventBroker(m)
	m.On("RemoveNode", mock.Anything, &pb.ContainerID{Id: "123456"}, mock.Anything).
		Return(&pb.Response{}, errors.New("error"))
	e := b.RemoveNode("123456")
	m.AssertNumberOfCalls(t, "RemoveNode", 1)
	assert.NotNil(t, e)
}
