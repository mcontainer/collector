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

func delta(previous bitarray.BitArray, current bitarray.BitArray) bitarray.BitArray {

	if previous.IsEmpty() && current.IsEmpty() {
		return current
	}

	if previous.IsEmpty() && !current.IsEmpty() {
		return current
	}

	if !previous.IsEmpty() && current.IsEmpty() {
		return current
	}

	if !previous.IsEmpty() && !current.IsEmpty() {
		return previous.And(current)
	}

	return current
}
