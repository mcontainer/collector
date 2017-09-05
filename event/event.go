package event

import (
	"context"
	pb "docker-visualizer/docker-graph-aggregator/events"
	log "github.com/sirupsen/logrus"
)

type NetworkEvent struct {
	IpSrc string
	IpDst string
	Size  uint16
}

type EventBroker struct {
	Stream *chan NetworkEvent
	grpc   pb.EventServiceClient
}

func NewEventBroker(client pb.EventServiceClient) *EventBroker {
	log.Info("Broker:: Start creating event")
	c := make(chan NetworkEvent)
	return &EventBroker{
		Stream: &c,
		grpc:   client,
	}
}

func (b *EventBroker) Listen() {

	stream, err := b.grpc.PushEvent(context.Background())
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
		if err := stream.Send(&pb.Event{IpSrc: v.IpSrc, IpDst: v.IpDst, Size: uint32(v.Size), Stack: "toto"}); err != nil {
			log.Warn("Failed during sending event")
		}
	}
}
