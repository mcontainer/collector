package namespace

import (
	"docker-visualizer/collector/event"
	"errors"
	"github.com/Workiva/go-datastructures/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"path/filepath"
	"testing"
)

type FakeNamespace struct {
	mock.Mock
}

func (n *FakeNamespace) findNetworkNamespace(rootPath string, namespace string) (result string, err error) {
	args := n.Called(rootPath, namespace)
	return args.String(0), args.Error(1)
}

func (n *FakeNamespace) runInNamespace(path string, node string, broker *event.EventBroker, wait *chan struct{}) error {
	args := n.Called(path, node, broker, wait)
	return args.Error(0)
}

func TestNamespace_IsRunningTrue(t *testing.T) {
	nspace := NewNamespace()
	nspace.GetRunningNetworks().Add("123")
	running := nspace.IsRunning("123")
	assert.Equal(t, true, running)
}

func TestNamespace_IsRunningFalse(t *testing.T) {
	nspace := NewNamespace()
	running := nspace.IsRunning("123")
	assert.Equal(t, false, running)
}

func TestNamespace_RunSuccess(t *testing.T) {
	mockNamespace := FakeNamespace{}
	namespace := Namespace{
		isRunning: set.New(),
		helper:    &mockNamespace,
	}
	ns := "1-123456789a"
	path := "/var/run/docker/netns/"
	wait := make(chan struct{})
	completePath := filepath.Join(path, ns)
	mockNamespace.
		On("findNetworkNamespace", path, ns).Return(ns, nil).
		On("runInNamespace", completePath, "toto", &event.EventBroker{}, &wait).Return(nil).
		Run(func(args mock.Arguments) {
			wait <- struct{}{}
		})
	e := namespace.Run(ns, "toto", &event.EventBroker{}, &wait)
	assert.Empty(t, e)
	assert.Equal(t, int64(1), namespace.isRunning.Len())
	assert.Equal(t, true, namespace.IsRunning(ns))
}

func TestNamespace_RunFailure(t *testing.T) {
	mockNamespace := FakeNamespace{}
	namespace := Namespace{
		isRunning: set.New(),
		helper:    &mockNamespace,
	}
	ns := "1-123456789a"
	path := "/var/run/docker/netns/"
	wait := make(chan struct{})
	completePath := filepath.Join(path, ns)
	mockNamespace.
		On("findNetworkNamespace", path, ns).Return(ns, nil).
		On("runInNamespace", completePath, "toto", &event.EventBroker{}, &wait).Return(errors.New("error"))
	e := namespace.Run(ns, "toto", &event.EventBroker{}, &wait)
	assert.NotEmpty(t, e)
}

func TestNamespace_RunFailure2(t *testing.T) {
	mockNamespace := FakeNamespace{}
	namespace := Namespace{
		isRunning: set.New(),
		helper:    &mockNamespace,
	}
	ns := "1-123456789a"
	path := "/var/run/docker/netns/"
	wait := make(chan struct{})
	mockNamespace.On("findNetworkNamespace", path, ns).Return(ns, errors.New("error"))
	e := namespace.Run(ns, "toto", &event.EventBroker{}, &wait)
	assert.NotEmpty(t, e)
}
