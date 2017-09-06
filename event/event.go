package event

import (
	"context"
	pb "docker-visualizer/proto/containers"
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

func (b *EventBroker) SendNode(container *types.Container) {
	ctx := context.Background()
	ip := ""
	networkName := ""
	for k, v := range container.NetworkSettings.Networks {
		networkName = k
		ip = v.IPAMConfig.IPv4Address
	}
	log.WithFields(log.Fields{
		"id":           container.ID,
		"name":         container.Names[0],
		"service":      container.Labels["com.docker.swarm.service.name"],
		"ip":           ip,
		"network name": networkName,
	}).Info("Broker:: Push container")
	r, e := b.grpc.AddNode(ctx, &pb.ContainerInfo{
		Id:      container.ID,
		Name:    container.Names[0],
		Service: container.Labels["com.docker.swarm.service.name"],
		Ip:      ip,
		Network: networkName,
		Stack:   "microservice",
	})
	if e != nil {
		log.WithField("Error", e).Fatal("Broker:: An error occured while sending grpc request")
	}
	if !r.Success {
		log.Warn("Broker:: The node " + container.ID + " has not been added to th database")
	}
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
		if err := stream.Send(&pb.ContainerEvent{IpSrc: v.IpSrc, IpDst: v.IpDst, Size: uint32(v.Size), Stack: "toto"}); err != nil {
			log.Warn("Failed during sending event")
		}
	}
}
