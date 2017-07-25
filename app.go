package main

import (
	"flag"
	"log"
	"docker-visualizer/docker-event-collector/docker"
	"docker-visualizer/docker-event-collector/packetbeat"
	"docker-visualizer/docker-event-collector/event"
	"fmt"
)

var (
	configPath    = flag.String("config_path", "[to be defined]", "docker config file path")
	httpPort            = flag.String("http_port", "5000", "http port to sending events")
	VERSION    string
	COMMIT     string
	BRANCH     string
)

const (
	PROTOCOL = "http"
	MODE     = "bridge"
)

func main() {

	flag.Parse()

	fmt.Printf("VERSION: %s - COMMIT: %s - BRANCH: %s \n", VERSION, COMMIT, BRANCH)

	if *configPath == "[to be defined]" {
		log.Fatal("A Packetbeat file path must be define")
	}

	d := docker.NewDockerClient()
	f := packetbeat.NewConfigFile(*configPath)
	b := event.NewEventBroker()

	go b.Listen()
	d.Listener()

	for {
		container := <-d.Data
		settings := container.NetworkSettings.Networks[MODE]
		event := event.DockerEvent{
			ID:        container.ID,
			IpAddress: settings.IPAddress,
			Ports:     container.Ports,
		}
		portList := f.GetPortList(PROTOCOL)
		updatePort := f.UpdatePort(container, portList)
		f.WritePort(updatePort)
		b.In <- event
	}

}
