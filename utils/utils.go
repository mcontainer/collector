package utils

import (
	"os/exec"
	log "github.com/sirupsen/logrus"
	"bufio"
	"path/filepath"
	"os"
	"strings"
	"io"
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

func FindNetworkNamespace(rootPath string, namespace string) (result string, err error) {
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(info.Name(), namespace[:10]) {
			result = info.Name()
			return io.EOF
		}
		return nil
	})
	if err == io.EOF {
		err = nil
	}
	return result, err
}
