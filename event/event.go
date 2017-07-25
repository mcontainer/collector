package event

import (
	"github.com/docker/docker/api/types"
	"log"
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
	log.Println("Start creating event broker")
	c := make(chan DockerEvent)
	return &EventBroker{
		In: c,
	}
}

func (broker *EventBroker) Listen() {

	for {
		fmt.Println(<-broker.In)
	}
	fmt.Println("stop listening")
}
