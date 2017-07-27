package utils

import (
	"github.com/Workiva/go-datastructures/bitarray"
	"github.com/docker/docker/api/types"
)

func ToBitArray(array []types.Port) bitarray.BitArray {
	b := bitarray.NewBitArray(65535)
	for _, p := range array {
		b.SetBit(uint64(p.PrivatePort))
	}
	return b
}
