package jail

import (
	"sync"
	"testing"

	"github.com/Sirupsen/logrus"
	"gopkg.in/alexcesaro/statsd.v2"
)

// TestNewipService
func TestNewipService(t *testing.T) {
	ipSvc, err := NewipService(testConf, &logrus.Logger{}, &statsd.Client{})
	if err != nil {
		t.Error("expected err to be nil")
	}
	if ipSvc == nil {
		t.Error("expected not nil ip service")
	}
}

// TestPopulatePool
func TestPopulatePool(t *testing.T) {
	ipSvc := &ipService{
		logger:  &logrus.Logger{},
		conf:    testConf,
		metrics: &statsd.Client{},
		mu:      &sync.Mutex{},
		pool:    make(map[string]byte),
	}
	ipSvc.populatePool()
	poolSize := len(ipSvc.pool)
	if poolSize != 201 {
		t.Errorf("expected %d got %d", 201, poolSize)
	}
}

// TestAllocate
func TestAllocate(t *testing.T) {}
