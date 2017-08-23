package sniffer

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/google/gopacket/tcpassembly"
	"bufio"
	"net/http"
	"io"
	log "github.com/sirupsen/logrus"
	"github.com/google/gopacket/examples/util"
	"github.com/google/gopacket/pcap"
	"time"
	"github.com/google/gopacket/layers"
)

//
//var iface = flag.String("i", "any", "Interface to get packets from")
//var snaplen = flag.Int("s", 1600, "SnapLen for pcap packet capture")
//var filter = flag.String("f", "tcp and dst port 80", "BPF filter for pcap")
//var node = flag.String("n", "", "Id of the current node")

type httpStreamFactory struct{}

type httpStream struct {
	net, transport gopacket.Flow
	reader         tcpreader.ReaderStream
}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	stream := &httpStream{
		net:       net,
		transport: transport,
		reader:    tcpreader.NewReaderStream(),
	}
	go stream.run()
	return &stream.reader
}

func (h *httpStream) run() {
	buffer := bufio.NewReader(&h.reader)
	for {
		req, err := http.ReadRequest(buffer)
		if err == io.EOF {
			return
		} else if err != nil {
			log.WithFields(log.Fields{
				"net":       h.net,
				"transport": h.transport,
				"error":     err,
			}).Fatal("Error reading stream")
		} else {
			bytes := tcpreader.DiscardBytesToEOF(req.Body)
			req.Body.Close()
			log.WithFields(log.Fields{
				"net":       h.net,
				"transport": h.transport,
				"req":       req,
				"bytes":     bytes,
			}).Info("Received request from stream")
		}
	}
}

func Capture(iface string, snaplen int32, filter string, node string, pipe *chan string) {
	defer util.Run()()
	var handle *pcap.Handle
	var err error

	log.WithField("interface", iface).Info("Starting capture")

	handle, err = pcap.OpenLive(iface, snaplen, true, pcap.BlockForever)

	if err != nil {
		log.Fatal(err)
	}

	if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatal(err)
	}

	streamFactory := &httpStreamFactory{}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	log.Info("Reading in packets")

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packets := packetSource.Packets()
	ticker := time.Tick(time.Minute)

	for {
		select {
		case packet := <-packets:
			if packet == nil {
				log.Info("Nil packet")
				return
			}
			tcp := packet.TransportLayer().(*layers.TCP)

			log.WithFields(log.Fields{
				"ip src":   packet.NetworkLayer().NetworkFlow().Src().String(),
				"ip dst":   packet.NetworkLayer().NetworkFlow().Dst().String(),
				"port src": tcp.SrcPort,
				"port dst": tcp.DstPort,
				"node":     node,
			}).Info("capture traffic")
			*pipe <- packet.NetworkLayer().NetworkFlow().String()

			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
				log.Info("Unusable packet")
				continue
			}
		case <-ticker:
			log.Info("Flush")
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}

}
