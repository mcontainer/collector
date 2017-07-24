package main

import (
	"flag"
	"log"
	"docker-visualizer/docker-event-collector/docker"
	"docker-visualizer/docker-event-collector/packetbeat"
)

var (
	packetbeatFilePath = flag.String("config_path", "[to be defined]", "docker config file path")
	httpPort           = flag.String("http_port", "5000", "http port to sending events")
)

const (
	PROTOCOL = "http"
)

func main() {

	flag.Parse()

	if *packetbeatFilePath == "[to be defined]" {
		log.Fatal("A Packetbeat file path must be define")
	}

	d := docker.NewDockerClient()
	f := packetbeat.NewConfigFile(*packetbeatFilePath)

	d.Listener()

	for {
		container := <-d.Data
		portList := f.GetPortList(PROTOCOL)
		updatePort := f.UpdatePort(container, portList)
		f.WritePort(updatePort)
	}

}
