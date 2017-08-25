package event

import (
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

type DockerEvent struct {
	ID        string
	IpAddress string
	Ports     []types.Port
}

type EventBroker struct {
	In *chan string
}

func NewEventBroker(c *chan string) *EventBroker {
	log.Info("Start creating event broker")
	return &EventBroker{
		In: c,
	}
}

func (broker *EventBroker) Listen() {
	for {
		log.WithField("detail", <-*broker.In).Info("NETWORK TRAFFIC")
	}
}
