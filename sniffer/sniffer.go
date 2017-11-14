package sniffer

import (
	"docker-visualizer/collector/event"
	"docker-visualizer/collector/sniffer/mmap"
	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
	"regexp"
)

const (
	HTTP_REGEXP = `HTTP\/\d\.\d\s+(\d+)\s+.*`
)

var (
	ip4        layers.IPv4
	eth        layers.Ethernet
	ip6        layers.IPv6
	tcp        layers.TCP
	payload    gopacket.Payload
	httpRegexp = regexp.MustCompile(HTTP_REGEXP)
)

func Capture(device string, node string, pipe *chan event.NetworkEvent, lock *chan struct{}) error {
	defer util.Run()()
	decodedLayers := make([]gopacket.LayerType, 0, 5)

	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip4, &ip6, &tcp, &payload)

	h, err := mmap.New(device, afpacket.DefaultPollTimeout)
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
			foundNetLayer := false
			for _, typ := range decodedLayers {
				switch typ {
				case layers.LayerTypeIPv4:
					foundNetLayer = true
					log.Info("Sniffer:: Successfully decoded layer type Ipv4")
					*pipe <- event.NetworkEvent{
						IpSrc: ip4.SrcIP.String(),
						IpDst: ip4.DstIP.String(),
						Size:  ip4.Length - uint16(ip4.IHL*4),
					}
				case layers.LayerTypeTCP:
					if foundNetLayer {
						log.Info("Application layer/Payload found attach to " + ip4.NetworkFlow().String())
						s := string(payload.Payload())
						codes := httpRegexp.FindStringSubmatch(s)
						if len(codes) >= 2 {
							log.WithField("code", codes[1]).Info("Sniffer:: Http status find")
						}
					} else {
						log.Warn("Counld not find IPv4 layer")
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
