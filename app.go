package main

import (
	"flag"
	"docker-visualizer/docker-event-collector/docker"
	log "github.com/sirupsen/logrus"
	"github.com/lstoll/cni/pkg/ns"
	"docker-visualizer/docker-sniffer"
	"docker-visualizer/docker-event-collector/utils"
	"path/filepath"
)

var (
	VERSION    string
	COMMIT     string
	BRANCH     string
)

const (
	NS_PATH  = "/var/run/docker/netns/"
)

func main() {

	log.WithFields(log.Fields{
		"version": VERSION,
		"commit":  COMMIT,
		"branch":  BRANCH,
	}).Info("Starting collector")


	dockerClient := docker.NewDockerClient()

	go dockerClient.ListenSwarm()

	isRunning := make(map[string]bool)

	channel := make(chan string)

	for {
		select {
		case id := <-dockerClient.NetworkID:
			if !isRunning[id] {
				namespace, err := utils.FindNetworkNamespace(NS_PATH, id)
				if err != nil {
					log.WithField("Error", err.Error()).Fatal("Unable to find network namespace")
				}
				go ns.WithNetNSPath(filepath.Join(NS_PATH, namespace), func(netns ns.NetNS) error {
					docker_sniffer.Capture("any", 1024, "tcp", "test_node", &channel)
					return nil
				})
				isRunning[id] = true
			} else {
				log.Info("Network already monitored")
			}

		case msg := <-channel:
			log.WithField("detail", msg).Info("NETWORK TRAFFIC")
		}
	}

}
