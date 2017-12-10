package jail

import (
	"errors"
	"net"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/briandowns/sky-island/config"
	"gopkg.in/alexcesaro/statsd.v2"
)

// IPServicer defines the behavior of the IP service
type IPServicer interface {
	Allocate() (string, error)
	Pool() map[string]byte
	UpdateIPState(string, byte) error
}

// ipService holds the state of the service
type ipService struct {
	logger  *logrus.Logger
	conf    *config.Config
	metrics *statsd.Client
	mu      sync.Locker
	pool    map[string]byte
}

// NewipService creates a new value of type ipService pointer
func NewipService(conf *config.Config, l *logrus.Logger, metrics *statsd.Client) (IPServicer, error) {
	i := ipService{
		logger:  l,
		conf:    conf,
		metrics: metrics,
		mu:      &sync.Mutex{},
		pool:    make(map[string]byte),
	}
	if err := i.populatePool(); err != nil {
		return nil, err
	}
	return &i, nil
}

// populatePool takes the given IP range from configuration and
// adds the IP addresses to the pool for allocation
func (i *ipService) populatePool() error {
	t := i.metrics.NewTiming()
	defer t.Send("populate_pool")
	ip := net.ParseIP(i.conf.IP4.StartAddr)
	ip = ip.To4()
	if ip == nil {
		return errors.New("bad start IP provided in config")
	}
	i.pool[ip.String()] = 0
	for j := ip[3]; int(j) < i.conf.IP4.Range; j++ {
		ip[3]++
		i.pool[ip.String()] = 0
	}
	return nil
}

// Allocate checks for available ip addresses returns one
// if available
func (i *ipService) Allocate() (string, error) {
	t := i.metrics.NewTiming()
	defer t.Send("allocate")
	i.mu.Lock()
	defer i.mu.Unlock()
	for k := range i.pool {
		if i.pool[k] == 0 {
			i.pool[k] = 1
			i.metrics.Histogram(k, 1)
			return k, nil
		}
	}
	return "", errors.New("no addresses available")
}

// Return
func (i *ipService) Return(ip string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.pool[ip] = 0
}

// Pool returns the current state of the IP address
// pool
func (i *ipService) Pool() map[string]byte {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.pool
}

// ValidIPState checks if the given state is valid
func ValidIPState(s byte) bool {
	return s == 0 || s == 1
}

// UpdateIPState iterates through the pool of IP addresses
// and if found sets it to the given state
func (i *ipService) UpdateIPState(ip string, state byte) error {
	t := i.metrics.NewTiming()
	defer t.Send("update_ip_state")
	i.mu.Lock()
	defer i.mu.Unlock()
	for k := range i.pool {
		if k == ip {
			i.pool[k] = state
			return nil
		}
	}
	return errors.New("unknown ip")
}
