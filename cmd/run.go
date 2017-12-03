package cmd

import (
	"docker-visualizer/collector/connector"
	"docker-visualizer/collector/connector/docker"
	"docker-visualizer/collector/event"
	"docker-visualizer/collector/namespace"
	"docker-visualizer/collector/util"
	pb "docker-visualizer/proto/containers"
	"github.com/cenkalti/backoff"
	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func createRunCmd(name string) *cobra.Command {
	var verbose bool
	var aggregator string
	var node string
	runCmd := cobra.Command{
		Use:   "run",
		Short: "Run " + name,
		Run: func(cmd *cobra.Command, args []string) {

			if !util.IsRoot() {
				log.Fatal("Please run this program as root !")
			}

			if verbose {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetLevel(log.WarnLevel)
			}
			ctx := context.Background()
			hostname := util.FindHostname()
			util.SetAggregatorEndpoint(&aggregator)
			conn, err := grpc.Dial(aggregator, grpc.WithInsecure())
			if err != nil {
				log.WithField("Error", err).Fatal("Error while creating grpc connection")
			}
			grpcClient := pb.NewContainerServiceClient(conn)
			defer conn.Close()

			client := docker.NewDockerClient()
			nspace := namespace.NewNamespace()
			fetcher := connector.NewFetcher(client)
			broker := event.NewEventBroker(grpcClient)

			networks, err := fetcher.FindOverlayNetworks()
			if err != nil {
				log.WithField("Error", err).Warn("App:: Overlay networks")
			}
			for _, network := range networks {
				if network.Name != "ingress" {
					wait := make(chan struct{})
					err := nspace.Run(network.ID, node, broker, &wait)
					if err != nil {
						log.WithField("Error", err).Warn("App:: Error while processing event")
					} else {
						containers, err := fetcher.DockerFromNetwork(network.ID)
						if err != nil {
							log.WithField("Error", err).Error("App:: Error while retrieving containers")
						}
						for _, c := range containers {
							if e := broker.SendNode(c, network.Name, hostname); e != nil {
								log.WithField("Error", e).Error("Broker:: An error occured while sending node")
							}
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
					case connector.ACTION_CONNECT:
						if !nspace.IsRunning(info.NetworkId) {
							wait := make(chan struct{})
							if err := nspace.Run(info.NetworkId, node, broker, &wait); err != nil {
								log.WithField("Error", err).Fatal("App:: Error while processing event")
							}
						} else {
							log.WithField("id", info.NetworkId).Info("App:: Network already monitored")
						}
						containerChan := make(chan *types.Container)
						operation := func() error {
							c, e := fetcher.FilterContainer(ctx, info.ContainerId)
							if c != nil {
								containerChan <- c
							}
							return e
						}
						go backoff.Retry(operation, backoff.NewExponentialBackOff())
						container := <-containerChan
						if err != nil {
							log.Error(err)
						} else {
							if e := broker.SendNode(container, info.NetworkName, hostname); e != nil {
								log.WithField("Error", e).Warn("Broker:: An error occured while sending node")
							}
						}
					case connector.ACTION_DISCONNECT:
						if e := broker.RemoveNode(info.ContainerId); e != nil {
							log.WithField("Error", e).Warn("Broker:: An error occured while removing node")
						}
					}
				case err := <-netErrors:
					log.WithField("Error", err).Error("App:: An error occured on events stream")
				}
			}

		},
	}
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	runCmd.Flags().StringVarP(&aggregator, "aggregator", "a", "127.0.0.1:10000", "aggregator endpoint")
	runCmd.Flags().StringVarP(&node, "node", "n", "sample-name", "node name")
	return &runCmd
}
