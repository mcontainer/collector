package main

import (
	"flag"
	"docker-visualizer/docker-event-collector/docker"
	"docker-visualizer/docker-event-collector/packetbeat"
	"docker-visualizer/docker-event-collector/event"
	log "github.com/sirupsen/logrus"
	"docker-visualizer/docker-event-collector/utils"
	"fmt"
	"github.com/lstoll/cni/pkg/ns"
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

		dockerClient := docker.NewDockerClient()

		go dockerClient.ListenSwarm()

		isRunning := make(map[string]bool)

		println("Start listening docker socket")

		channel := make(chan string)

		for {
			select {
			case id := <-dockerClient.NetworkID:
				fmt.Println(id)
				if !isRunning[id] {
					path := "/run/docker/netns/"
					worker := &utils.Worker{
						Command: "./sniffer.sh",
						Args:    append([]string{path}),
						Output:  &channel,
					}

					go ns.WithNetNSPath("/var/run/docker/netns/1-rm7o5uxjmo", func(netns ns.NetNS) error {
						log.Info("path = " + netns.Path())
						worker.Run()
						return nil
					})
					isRunning[id] = true
				} else {
					log.Info("Network already monitored")
				}

				//go worker.Run()
			case msg := <-channel:
				log.Info(msg)
			}
		}

		//workerMap := make(map[string]int)

		//c := exec.Command(worker.Command, worker.Args...)
		//ch := make(chan struct{})
		//go worker.Run()
		//ch <- struct{}{}
		//c.Start()


		//<-ch
		//if err := c.Wait(); err != nil {
		//	fmt.Println(err)
		//}
		//fmt.Println("done.")

		//for {
		//	select {
		//	case pid := <-worker.Pid:
		//		log.Info(pid)
		//	}
		//}

		//args := append([]string{"--net=" + path + "1-rm7o5uxjmo"}, "tcpdump", "-i", "any", "tcp")
		//c := exec.Command("nsenter", args...)
		//c.Stderr = os.Stderr
		//stdout, _ := c.StdoutPipe()
		//scan := bufio.NewScanner(stdout)
		//
		//go func() {
		//	for scan.Scan() {
		//		fmt.Println(scan.Text())
		//	}
		//}()
		//
		//c.Start()
		//
		//err := c.Wait()
		//if err != nil {
		//	fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		//	return
		//}

	}

}
