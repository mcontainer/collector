package utils

import (
	"github.com/Workiva/go-datastructures/bitarray"
	"github.com/docker/docker/api/types"
	"os/exec"
	log "github.com/sirupsen/logrus"
	"bufio"
)

type Worker struct {
	Command string
	Args    []string
	Output  *chan string
	Pid     chan int
}

func (cmd *Worker) Run() {
	c := exec.Command(cmd.Command, cmd.Args...)
	stdout, err := c.StdoutPipe()
	if err != nil {
		panic(err)
	}
	b := bufio.NewScanner(stdout)
	go func() {
		for b.Scan() {
			log.Info(b.Text())
			*cmd.Output <- b.Text()
		}
	}()
	log.Info("Start command")
	c.Start()
	cmd.Pid <- c.Process.Pid
}

func CmdCollector(channel *chan string) {
	for {
		msg := <-*channel
		log.Info(msg)
	}
}

func ToBitArray(array []types.Port) bitarray.BitArray {
	b := bitarray.NewBitArray(65535)
	for _, p := range array {
		b.SetBit(uint64(p.PrivatePort))
	}
	return b
}
