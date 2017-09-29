package main

import (
	"context"
	"docker-visualizer/docker-event-collector/docker"
	"docker-visualizer/docker-event-collector/event"
	"docker-visualizer/docker-event-collector/namespace"
	pb "docker-visualizer/proto/containers"
	"flag"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
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
	ctx := context.Background()
	endpoint := os.Getenv("AGGREGATOR")
	if endpoint != "" {
		log.WithField("aggregator", endpoint).Info("Find env variable")
		aggregator = &endpoint
	}
	conn, err := grpc.Dial(*aggregator, grpc.WithInsecure())
	if err != nil {
		log.WithField("Error", err).Fatal("Error while creating grpc connection")
	}
	grpcClient := pb.NewContainerServiceClient(conn)
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
			err := nspace.Run(network.ID, *node, broker)
			if err != nil {
				log.WithField("Error", err).Warn("App:: Error while processing event")
			} else {
				containers, err := fetcher.DockerFromNetwork(network.ID)
				if err != nil {
					log.WithField("Error", err).Warn("App:: Error while retrieving containers")
				}
				for _, c := range containers {
					broker.SendNode(c, network.Name)
				}
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
			switch info.Action {
			case docker.ACTION_CONNECT:
				if !nspace.IsRunning.Exists(info.NetworkId) {
					if err := nspace.Run(info.NetworkId, *node, broker); err != nil {
						log.WithField("Error", err).Fatal("App:: Error while processing event")
					}
				} else {
					log.WithField("id", info.NetworkId).Info("App:: Network already monitored")
				}
				container, err := fetcher.FilterContainer(ctx, info.ContainerId)
				if err != nil {
					log.Fatal(err)
				}
				broker.SendNode(container, info.NetworkName)
			case docker.ACTION_DISCONNECT:
				broker.RemoveNode(info.ContainerId)
			}
		case err := <-netErrors:
			log.WithField("Error", err).Fatal("App:: An error occured on events stream")
		}
	}

}
