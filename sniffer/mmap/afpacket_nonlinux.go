// +build !linux

package mmap

import "errors"

type afpacketSniffer struct{}

const AfpacketErrorSystem = errors.New("Afpacket MMAP sniffing is only available on Linux")

func newAfpacketSniffer(device string, timeout time.Duration) (h *afpacketSniffer, err error) {
	return nil, AfpacketErrorSystem
}

func (h *afpacketSniffer) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	return data, ci, AfpacketErrorSystem
}

func (h *afpacketSniffer) Close() {
}
