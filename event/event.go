package event

import (
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"fmt"
)

type DockerEvent struct {
	ID        string
	IpAddress string
	Ports     []types.Port
}

type EventBroker struct {
	In chan DockerEvent
}

func NewEventBroker() *EventBroker {
	log.Info("Start creating event broker")
	c := make(chan DockerEvent)
	return &EventBroker{
		In: c,
	}
}

func (broker *EventBroker) Listen() {
	go func() {
		for {
			fmt.Println(<-broker.In)
		}
	}()
}
