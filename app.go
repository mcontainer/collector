package main

import (
	"flag"
	"docker-visualizer/docker-event-collector/docker"
	"docker-visualizer/docker-event-collector/packetbeat"
	"docker-visualizer/docker-event-collector/event"
	log "github.com/sirupsen/logrus"
	"docker-visualizer/docker-event-collector/utils"
)

var (
	configPath = flag.String("config_path", "[to be defined]", "docker config file path")
	httpPort   = flag.String("http_port", "5000", "http port to sending events")
	VERSION    string
	COMMIT     string
	BRANCH     string
)

const (
	PROTOCOL  = "http"
	MODE      = "bridge"
	MAX_PORTS = 65535
)

func main() {

	flag.Parse()

	log.WithFields(log.Fields{
		"version": VERSION,
		"commit":  COMMIT,
		"branch":  BRANCH,
	}).Info("Starting collector")

	if *configPath == "[to be defined]" {
		log.Fatal("A Packetbeat file path must be define")
	}

	dockerClient := docker.NewDockerClient()
	config := packetbeat.NewConfigFile(*configPath)
	broker := event.NewEventBroker()

	broker.Listen()
	dockerClient.Listen()

	for {
		container := <-dockerClient.Data
		settings := container.NetworkSettings.Networks[MODE]
		event := event.DockerEvent{
			ID:        container.ID,
			IpAddress: settings.IPAddress,
			Ports:     container.Ports,
		}
		bitPortsContainer := utils.ToBitArray(container.Ports)
		bitPortList := config.GetPortList(PROTOCOL)
		bitPorts, shouldWrite := config.UpdatePort(bitPortsContainer, bitPortList)
		if shouldWrite {
			config.WritePort(bitPorts)
		}
		broker.In <- event
	}

}
