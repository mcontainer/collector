package main

import (
	"docker-visualizer/collector/cmd"
	"flag"
	log "github.com/sirupsen/logrus"
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

	if e := cmd.CreateRootCmd("collector", VERSION, COMMIT, BRANCH).Execute(); e != nil {
		log.Error(e)
		os.Exit(1)
	}

	//time.Sleep(10 * time.Minute)
	//
	//flag.Parse()
	//
	//ctx := context.Background()
	//hostname := util.FindHostname()
	//util.SetAggregatorEndpoint(aggregator)
	//conn, err := grpc.Dial(*aggregator, grpc.WithInsecure())
	//if err != nil {
	//	log.WithField("Error", err).Fatal("Error while creating grpc connection")
	//}
	//grpcClient := pb.NewContainerServiceClient(conn)
	//defer conn.Close()
	//
	//client := docker.NewDockerClient()
	//nspace := namespace.NewNamespace()
	//fetcher := connector.NewFetcher(client)
	//broker := event.NewEventBroker(grpcClient)
	//
	//networks, err := fetcher.FindOverlayNetworks()
	//if err != nil {
	//	log.WithField("Error", err).Warn("App:: Overlay networks")
	//}
	//for _, network := range networks {
	//	if network.Name != "ingress" {
	//		wait := make(chan struct{})
	//		err := nspace.Run(network.ID, *node, broker, &wait)
	//		if err != nil {
	//			log.WithField("Error", err).Warn("App:: Error while processing event")
	//		} else {
	//			containers, err := fetcher.DockerFromNetwork(network.ID)
	//			if err != nil {
	//				log.WithField("Error", err).Error("App:: Error while retrieving containers")
	//			}
	//			for _, c := range containers {
	//				if e := broker.SendNode(c, network.Name, hostname); e != nil {
	//					log.WithField("Error", e).Error("Broker:: An error occured while sending node")
	//				}
	//			}
	//		}
	//	} else {
	//		fetcher.IngressId = network.ID
	//		log.WithField("ID", fetcher.IngressId).Info("App:: Find Ingress network id")
	//	}
	//}
	//
	//netEvents, netErrors := fetcher.ListenNetwork()
	//go broker.Listen()
	//
	//for {
	//	select {
	//	case info := <-netEvents:
	//		switch info.Action {
	//		case connector.ACTION_CONNECT:
	//			if !nspace.IsRunning(info.NetworkId) {
	//				wait := make(chan struct{})
	//				if err := nspace.Run(info.NetworkId, *node, broker, &wait); err != nil {
	//					log.WithField("Error", err).Fatal("App:: Error while processing event")
	//				}
	//			} else {
	//				log.WithField("id", info.NetworkId).Info("App:: Network already monitored")
	//			}
	//			container, err := fetcher.FilterContainer(ctx, info.ContainerId)
	//			if err != nil {
	//				log.Error(err)
	//			}
	//			if e := broker.SendNode(container, info.NetworkName, hostname); e != nil {
	//				log.WithField("Error", e).Warn("Broker:: An error occured while sending node")
	//			}
	//		case connector.ACTION_DISCONNECT:
	//			if e := broker.RemoveNode(info.ContainerId); e != nil {
	//				log.WithField("Error", e).Warn("Broker:: An error occured while removing node")
	//			}
	//		}
	//	case err := <-netErrors:
	//		log.WithField("Error", err).Error("App:: An error occured on events stream")
	//	}
	//}

}
