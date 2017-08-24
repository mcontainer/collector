package sniffer

import (
	"github.com/google/gopacket/afpacket"
	"time"
	"github.com/google/gopacket"
	log "github.com/sirupsen/logrus"
	"fmt"
	"os"
)

type afpacketSniffer struct {
	TPacket *afpacket.TPacket
}

//From packetbeat
func afpacketComputeSize(target_size_mb int, snaplen int, page_size int) (
	frame_size int, block_size int, num_blocks int, err error) {

	if snaplen < page_size {
		frame_size = page_size / (page_size / snaplen)
	} else {
		frame_size = (snaplen/page_size + 1) * page_size
	}

	// 128 is the default from the gopacket library so just use that
	block_size = frame_size * 128
	num_blocks = (target_size_mb * 1024 * 1024) / block_size

	if num_blocks == 0 {
		return 0, 0, 0, fmt.Errorf("Buffer size too small")
	}

	return frame_size, block_size, num_blocks, nil
}

func newAfpacketSniffer(device string, timeout time.Duration) (h *afpacketSniffer, err error) {

	const (
		buffer_mb int  = 24
		snaplen   int  = 65536
	)

	frameSize, blockSize, numBlocks, e := afpacketComputeSize(
		buffer_mb,
		snaplen,
		os.Getpagesize(),
	)
	if e != nil {
		log.WithField("error", e).Fatal("Error while calculating afpacket size")
	}

	h = &afpacketSniffer{}
	if device == "any" {
		h.TPacket, err = afpacket.NewTPacket(
			afpacket.OptFrameSize(frameSize),
			afpacket.OptBlockSize(blockSize),
			afpacket.OptNumBlocks(numBlocks),
			afpacket.OptPollTimeout(timeout),
		)
	} else {
		h.TPacket, err = afpacket.NewTPacket(
			afpacket.OptInterface(device),
			afpacket.OptFrameSize(snaplen),
			afpacket.OptBlockSize(blockSize),
			afpacket.OptNumBlocks(numBlocks),
			afpacket.OptPollTimeout(timeout),
		)
	}
	return h, err
}

func (h *afpacketSniffer) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	return h.TPacket.ZeroCopyReadPacketData()
}

func (h *afpacketSniffer) Close() {
	h.TPacket.Close()
}
