package jail

import (
	"sync"
	"testing"

	"github.com/Sirupsen/logrus"
	"gopkg.in/alexcesaro/statsd.v2"
)

// TestNewNetworkService
func TestNewNetworkService(t *testing.T) {
	networkSvc, err := NewNetworkService(testConf, &logrus.Logger{}, &statsd.Client{})
	if err != nil {
		t.Error("expected err to be nil")
	}
	if networkSvc == nil {
		t.Error("expected not nil ip service")
	}
}

// TestPopulatePool
func TestPopulatePool(t *testing.T) {
	networkSvc := &networkService{
		logger:  &logrus.Logger{},
		conf:    testConf,
		metrics: &statsd.Client{},
		mu:      &sync.Mutex{},
		ip4Pool: make(map[string]byte),
	}
	networkSvc.populatePool()
	poolSize := len(networkSvc.ip4Pool)
	if poolSize != 201 {
		t.Errorf("expected %d got %d", 201, poolSize)
	}
}

// TestAllocate
func TestAllocate(t *testing.T) {}
