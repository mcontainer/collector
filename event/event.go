package event

import (
	"context"
	pb "docker-visualizer/proto/containers"
	"errors"
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

type NetworkEvent struct {
	IpSrc string
	IpDst string
	Size  uint16
}

type EventBroker struct {
	Stream *chan NetworkEvent
	grpc   pb.ContainerServiceClient
}

func NewEventBroker(client pb.ContainerServiceClient) *EventBroker {
	log.Info("Broker:: Start creating event")
	c := make(chan NetworkEvent)
	return &EventBroker{
		Stream: &c,
		grpc:   client,
	}
}

func (b *EventBroker) SendNode(container *types.Container, network, hostname string) error {
	ctx := context.Background()
	ip := container.NetworkSettings.Networks[network].IPAMConfig.IPv4Address
	log.WithFields(log.Fields{
		"id":           container.ID,
		"name":         container.Names[0],
		"service":      container.Labels["com.docker.swarm.service.name"],
		"ip":           ip,
		"network name": network,
	}).Info("Broker:: Push container")
	r, e := b.grpc.AddNode(ctx, &pb.ContainerInfo{
		Id:      container.ID,
		Name:    container.Names[0],
		Service: container.Labels["com.docker.swarm.service.name"],
		Ip:      ip,
		Network: network,
		Stack:   "microservice",
		Host:    hostname,
	})
	if e != nil {
		return e
	}
	if !r.Success {
		return errors.New("The node " + container.ID + " has not been added to the database")
	}
	return nil
}

func (b *EventBroker) RemoveNode(containerID string) error {
	ctx := context.Background()
	log.WithField("id", containerID).Info("Broker:: Remove container")
	r, e := b.grpc.RemoveNode(ctx, &pb.ContainerID{Id: containerID})
	if e != nil {
		return e
	}
	if !r.Success {
		log.Warn("Broker:: The node " + containerID + " has not been removed from the database")
		return errors.New("The node " + containerID + " has not been removed from the database")
	}
	log.Info("Broker:: Node " + containerID + " has been removed")
	return nil
}

func (b *EventBroker) Listen() {
	stream, err := b.grpc.StreamContainerEvents(context.Background())
	if err != nil {
		log.WithField("Error", err).Fatal("Broker:: Unable to connect to aggregator endpoint")
	}
	for {
		v := <-*b.Stream
		log.WithFields(log.Fields{
			"src":  v.IpSrc,
			"dst":  v.IpDst,
			"size": v.Size,
		}).Info("Broker:: Receive network traffic")
		if err := stream.Send(&pb.ContainerEvent{IpSrc: v.IpSrc, IpDst: v.IpDst, Size: uint32(v.Size), Stack: "toto", Host: "test"}); err != nil {
			log.Warn("Failed during sending event")
		}
	}
}
