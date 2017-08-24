package main

import (
	"docker-visualizer/docker-event-collector/docker"
	log "github.com/sirupsen/logrus"
	"github.com/lstoll/cni/pkg/ns"
	"docker-visualizer/docker-event-collector/utils"
	"path/filepath"
	"docker-visualizer/docker-event-collector/sniffer"
)

var (
	VERSION string
	COMMIT  string
	BRANCH  string
)

const (
	NS_PATH = "/var/run/docker/netns/"
)

func main() {

	log.WithFields(log.Fields{
		"version": VERSION,
		"commit":  COMMIT,
		"branch":  BRANCH,
	}).Info("Starting collector")

	channel := make(chan string)

	dockerCLI := docker.NewDockerClient()

	go dockerCLI.Listen()

	isRunning := make(map[string]bool)


	for {
		select {
		case msg := <-dockerCLI.Data:
			if !isRunning[msg.NetworkId] {
				namespace, err := utils.FindNetworkNamespace(NS_PATH, msg.NetworkId)
				log.WithField("Network namespace", namespace).Info("Find Network Namespace")
				if err != nil {
					log.WithField("Error", err.Error()).Fatal("Unable to find network namespace")
				}
				log.WithField("path", filepath.Join(NS_PATH, namespace)).Info("Building Namespace path")
				go ns.WithNetNSPath(filepath.Join(NS_PATH, namespace), func(netns ns.NetNS) error {
					sniffer.Capture("any", "test_node", &channel)
					return nil
				})
				isRunning[msg.NetworkId] = true
			} else {
				log.Info("Network already monitored")
			}

		case msg := <-channel:
			log.WithField("detail", msg).Info("NETWORK TRAFFIC")
		}
	}

}
