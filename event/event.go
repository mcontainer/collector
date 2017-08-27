package event

import (
	log "github.com/sirupsen/logrus"
)

type NetworkEvent struct {
	IpSrc string
	IpDst string
	Size  uint16
}

type EventBroker struct {
	Stream *chan NetworkEvent
}

func NewEventBroker() *EventBroker {
	log.Info("Start creating event broker")
	c := make(chan NetworkEvent)
	return &EventBroker{
		Stream: &c,
	}
}

func (broker *EventBroker) Listen() {
	for {
		log.WithField("detail", <-*broker.Stream).Info("NETWORK TRAFFIC")
	}
}
