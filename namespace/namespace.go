package namespace

import (
	"context"
	"docker-visualizer/docker-event-collector/event"
	"docker-visualizer/docker-event-collector/sniffer"
	"errors"
	"github.com/Workiva/go-datastructures/set"
	"github.com/containernetworking/plugins/pkg/ns"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	NS_PATH = "/var/run/docker/netns/"
)

type Namespace struct {
	IsRunning *set.Set
}

func NewNamespace() *Namespace {
	return &Namespace{
		IsRunning: set.New(),
	}
}

func (nspace *Namespace) Run(networkID string, node string, broker *event.EventBroker) error {
	ctx, cancel := context.WithCancel(context.Background())
	wait := make(chan struct{})
	nsvalue, err := nspace.findNetworkNamespace(NS_PATH, networkID)
	if err != nil {
		log.WithField("Error", err).Warn("Namespace:: Unable to find network nsvalue")
		return err
	}
	log.WithField("id", nsvalue).Info("Namespace:: Find network nsvalue")
	path := filepath.Join(NS_PATH, nsvalue)
	log.WithField("path", path).Info("Namespace:: Building Namespace path")
	go func() {
		e := ns.WithNetNSPath(path, func(netNS ns.NetNS) error {
			if err := sniffer.Capture("any", node, broker.Stream, &wait); err != nil {
				log.WithField("Error", err).Fatal("Namespace:: Unable to start the sniffing process")
				return err
			}
			return nil
		})
		if e != nil {
			log.WithField("Error", err).Warn("Namespace:: Unable to enter in the network nsvalue")
			cancel()
		}
	}()

	select {
	case <-ctx.Done():
		return errors.New("Namespace:: an error occured")
	case <-wait:
		nspace.IsRunning.Add(networkID)
		log.WithField("id", networkID).Info("Namespace:: Current network is now monitored")
		return nil

	}

}

func (nspace *Namespace) findNetworkNamespace(rootPath string, namespace string) (result string, err error) {
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
