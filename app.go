package main

import (
	"docker-visualizer/docker-event-collector/docker"
	"docker-visualizer/docker-event-collector/event"
	"docker-visualizer/docker-event-collector/namespace"
	pb "docker-visualizer/proto/events"
	"flag"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	VERSION string
	COMMIT  string
	BRANCH  string
)

var (
	aggregator = flag.String("aggregator", "127.0.0.1:10000", "Endpoint to the aggregator service")
	node       = flag.String("node", "sample-node", "Specify node name")
)

func main() {
	flag.Parse()

	log.WithFields(log.Fields{
		"version": VERSION,
		"commit":  COMMIT,
		"branch":  BRANCH,
	}).Info("Starting collector")

	conn, err := grpc.Dial(*aggregator, grpc.WithInsecure())
	if err != nil {
		log.WithField("Error", err).Fatal("Error while creating grpc connection")
	}
	grpcClient := pb.NewEventServiceClient(conn)
	defer conn.Close()
	client := docker.NewDockerClient()
	nspace := namespace.NewNamespace()
	fetcher := docker.NewFetcher(client)
	broker := event.NewEventBroker(grpcClient)

	networks, err := fetcher.FindOverlayNetworks()
	if err != nil {
		log.WithField("Error", err).Warn("App:: Overlay networks")
	}
	for _, network := range networks {
		if network.Name != "ingress" {
			if err := nspace.Run(network.ID, *node, broker); err != nil {
				log.WithField("Error", err).Fatal("App:: Error while processing event")
			}
		} else {
			fetcher.IngressId = network.ID
			log.WithField("ID", fetcher.IngressId).Info("App:: Find Ingress network id")
		}
	}

	netEvents, netErrors := fetcher.ListenNetwork()
	go broker.Listen()

	for {
		select {
		case info := <-netEvents:
			if !nspace.IsRunning.Exists(info.NetworkId) {
				if err := nspace.Run(info.NetworkId, *node, broker); err != nil {
					log.WithField("Error", err).Fatal("App:: Error while processing event")
				}
			} else {
				log.WithField("id", info.NetworkId).Info("App:: Network already monitored")
			}
			//TODO: send node id to aggregator server through grpc
		case err := <-netErrors:
			log.WithField("Error", err).Fatal("App:: An error occured on events stream")
		}
	}

}
