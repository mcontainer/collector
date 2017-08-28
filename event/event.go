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
	log.Info("Start creating event b")
	c := make(chan NetworkEvent)
	return &EventBroker{
		Stream: &c,
		grpc:   client,
	}
}

func (b *EventBroker) Listen() {

	stream, err := b.grpc.PushEvent(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for {
		v := <-*b.Stream
		log.WithField("detail", v).Info("NETWORK TRAFFIC")
		if err := stream.Send(&pb.Event{IpSrc: v.IpSrc, IpDst: v.IpDst, Size: uint32(v.Size), Stack: "toto"}); err != nil {
			log.Warn("Failed during sending event")
		}
	}
}
