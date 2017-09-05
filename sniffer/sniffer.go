package sniffer

import (
	"docker-visualizer/docker-event-collector/event"
	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
)

//
//var iface = flag.String("i", "any", "Interface to get packets from")
//var snaplen = flag.Int("s", 1600, "SnapLen for pcap packet capture")
//var filter = flag.String("f", "tcp and dst port 80", "BPF filter for pcap")
//var node = flag.String("n", "", "Id of the current node")

var (
	ip4     layers.IPv4
	eth     layers.Ethernet
	ip6     layers.IPv6
	tcp     layers.TCP
	payload gopacket.Payload
)

func Capture(device string, node string, pipe *chan event.NetworkEvent, lock *chan struct{}) error {
	defer util.Run()()

	decodedLayers := make([]gopacket.LayerType, 0, 10)

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &payload)

	h, err := newAfpacketSniffer(device, afpacket.DefaultPollTimeout)
	if err != nil {
		log.WithField("error", err).Fatal("Sniffer:: Error while creating afpacket sniffer")
	}

	log.WithField("interface", device).Info("Sniffer:: Starting capture")
	dataSource := gopacket.NewPacketSource(h, layers.LayerTypeEthernet)
	packets := dataSource.Packets()
	*lock <- struct{}{}
	for {
		select {
		case packet := <-packets:
			err = parser.DecodeLayers(packet.Data(), &decodedLayers)
			for _, typ := range decodedLayers {
				switch typ {
				case layers.LayerTypeIPv4:
					log.Info("Sniffer:: Successfully decoded layer type Ipv4")
					*pipe <- event.NetworkEvent{
						IpSrc: ip4.SrcIP.String(),
						IpDst: ip4.DstIP.String(),
						Size:  ip4.Length - uint16(ip4.IHL*4),
					}
				}
			}
			if len(decodedLayers) == 0 {
				log.Warn("Sniffer:: Packet has been truncated")
			}
			if err != nil {
				log.WithField("err", err).Warn("Sniffer:: No decoder found")
			}
		}
	}

}
