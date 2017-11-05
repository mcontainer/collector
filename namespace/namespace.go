package namespace

import (
	"context"
	"docker-visualizer/collector/event"
	"docker-visualizer/collector/log"
	"docker-visualizer/collector/sniffer"
	"errors"
	"github.com/Workiva/go-datastructures/set"
	"github.com/containernetworking/plugins/pkg/ns"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	NS_PATH = "/var/run/docker/netns/"
)

type INamespace interface {
	Run(networkID string, node string, broker *event.EventBroker, wait *chan struct{}) error
	IsRunning(networkID string) bool
	GetRunningNetworks() *set.Set
}

type INamespaceHelper interface {
	findNetworkNamespace(rootPath string, namespace string) (result string, err error)
	runInNamespace(path string, node string, broker *event.EventBroker, wait *chan struct{}) error
}

type NamespaceHelper struct{}

type Namespace struct {
	isRunning *set.Set
	helper    INamespaceHelper
}

func NewNamespace() INamespace {
	return &Namespace{
		isRunning: set.New(),
		helper:    &NamespaceHelper{},
	}
}

func (nspace *Namespace) IsRunning(networkID string) bool {
	return nspace.isRunning.Exists(networkID)
}

func (nspace *Namespace) GetRunningNetworks() *set.Set {
	return nspace.isRunning
}

func (nspace *Namespace) Run(networkID string, node string, broker *event.EventBroker, wait *chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	name, err := nspace.helper.findNetworkNamespace(NS_PATH, networkID)
	if err != nil {
		log.WithField("Error", err).Warn("Namespace:: Unable to find network name")
		return err
	}
	path := filepath.Join(NS_PATH, name)
	log.WithField("path", path).Info("Namespace:: Building Namespace path")
	go func() {
		if nspace.helper.runInNamespace(path, node, broker, wait) != nil {
			log.WithField("Error", err).Warn("namespace:: Unable to enter in the network name")
			cancel()
		}
	}()
	select {
	case <-ctx.Done():
		return errors.New("namespace:: an error occured")
	case <-*wait:
		nspace.isRunning.Add(networkID)
		log.WithField("id", networkID).Info("namespace:: Current network is now monitored")
		return nil
	}

}

func (*NamespaceHelper) runInNamespace(path string, node string, broker *event.EventBroker, wait *chan struct{}) error {
	return ns.WithNetNSPath(path, func(netNS ns.NetNS) error {
		if err := sniffer.Capture("any", node, broker.Stream, wait); err != nil {
			log.WithField("Error", err).Fatal("Namespace:: Unable to start the sniffing process")
			return err
		}
		return nil
	})
}

func (*NamespaceHelper) findNetworkNamespace(rootPath string, namespace string) (result string, err error) {
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		name := info.Name()
		if strings.Contains(name, namespace[:10]) {
			result = name
			return io.EOF
		}
		return nil
	})
	if err == io.EOF {
		err = nil
	}
	return result, err
}
