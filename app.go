package main

import (
	"docker-visualizer/docker-event-collector/docker"
	log "github.com/sirupsen/logrus"
	"docker-visualizer/docker-event-collector/utils"
	"path/filepath"
	"docker-visualizer/docker-event-collector/sniffer"
	"docker-visualizer/docker-event-collector/event"
	"github.com/containernetworking/plugins/pkg/ns"
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

	client := docker.NewDockerClient()
	fetcher := docker.NewFetcher(client)
	broker := event.NewEventBroker()
	isRunning := make(map[string]bool)

	events, errors := fetcher.Listen()
	go broker.Listen()

	for {
		select {
		case msg := <-events:
			if !isRunning[msg.NetworkId] {
				namespace, err := utils.FindNetworkNamespace(NS_PATH, msg.NetworkId)
				log.WithField("Network namespace", namespace).Info("Find Network Namespace")
				if err != nil {
					log.WithField("Error", err.Error()).Fatal("Unable to find network namespace")
				}
				log.WithField("path", filepath.Join(NS_PATH, namespace)).Info("Building Namespace path")
				go func() {
					e := ns.WithNetNSPath(filepath.Join(NS_PATH, namespace), func(netns ns.NetNS) error {
						sniffer.Capture("any", "test_node", broker.Stream)
						return nil
					})
					if e != nil {
						log.WithField("error", e).Fatal("Error while entering in network namespace")
					}
				}()
				isRunning[msg.NetworkId] = true
			} else {
				log.WithField("network id", msg.NetworkId).Info("Network already monitored")
			}
		case err := <-errors:
			log.Fatal(err)
		}
	}

}
