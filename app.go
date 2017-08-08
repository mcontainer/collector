package main

import (
	"flag"
	"docker-visualizer/docker-event-collector/docker"
	"docker-visualizer/docker-event-collector/packetbeat"
	"docker-visualizer/docker-event-collector/event"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"os"
	"fmt"
	"bufio"
)

var (
	configPath = flag.String("config_path", "[to be defined]", "docker config file path")
	httpPort   = flag.String("http_port", "5000", "http port to sending events")
	swarm      = flag.String("swarm", "false", "set swarm mode")
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

	log.WithFields(log.Fields{
		"version": VERSION,
		"commit":  COMMIT,
		"branch":  BRANCH,
	}).Info("Starting collector")

	if *swarm == "false" {

		if *configPath == "[to be defined]" {
			log.Fatal("A Packetbeat file path must be define")
		}

		dockerClient := docker.NewDockerClient()
		config := packetbeat.NewConfigFile(*configPath)
		broker := event.NewEventBroker()

		go broker.Listen()
		go dockerClient.Listen()

		for {
			select {
			case container := <-dockerClient.Data:
				settings := container.NetworkSettings.Networks[MODE]
				event := event.DockerEvent{
					ID:        container.ID,
					IpAddress: settings.IPAddress,
					Ports:     container.Ports,
				}
				bitPortsContainer := dockerClient.ToBitPorts(container)
				bitPortList := config.GetPortList(PROTOCOL)
				bitPorts, shouldWrite := docker.UpdatePort(bitPortsContainer, bitPortList)
				if shouldWrite {
					config.WritePort(bitPorts)
				}
				broker.In <- event

			case idRemoved := <-dockerClient.Stop:
				bitPortList := config.GetPortList(PROTOCOL)
				updated := dockerClient.RemovePorts(idRemoved, bitPortList)
				config.WritePort(updated)
			}
		}
	} else {

		println("Swarm mode enabled")

		//dockerClient := docker.NewDockerClient()
		//
		//go dockerClient.ListenSwarm()
		//
		//println("Start listening docker socket")
		//
		//for {
		//	select {
		//	case container := <-dockerClient.Data:
		//		fmt.Println(container)
		//	}
		//}
		path := "/run/docker/netns/"
		args := append([]string{"--net=" + path + "1-rm7o5uxjmo"}, "tcpdump", "-i", "any", "tcp")
		c := exec.Command("nsenter", args...)
		c.Stderr = os.Stderr
		stdout, _ := c.StdoutPipe()
		scan := bufio.NewScanner(stdout)

		go func() {
			for scan.Scan() {
				fmt.Println(scan.Text())
			}
		}()

		c.Start()

		err := c.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
			return
		}

	}

}
