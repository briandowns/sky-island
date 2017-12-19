package jail

import (
	"errors"
	"net"
	"sync"

	"github.com/briandowns/sky-island/config"
	gklog "github.com/go-kit/kit/log"
	"gopkg.in/alexcesaro/statsd.v2"
)

// NetworkServicer defines the behavior of the IP service
type NetworkServicer interface {
	Allocate([]byte) (string, error)
	Pool() map[string][]byte
	UpdateIPState(string, []byte) error
}

// ipService holds the state of the service
type networkService struct {
	logger  gklog.Logger
	conf    *config.Config
	metrics *statsd.Client
	mu      sync.Locker
	ip4Pool map[string][]byte
}

// NewNetworkService creates a new value of type networkService pointer
func NewNetworkService(conf *config.Config, l gklog.Logger, metrics *statsd.Client) (NetworkServicer, error) {
	n := networkService{
		logger:  l,
		conf:    conf,
		metrics: metrics,
		mu:      &sync.Mutex{},
		ip4Pool: make(map[string][]byte),
	}
	if err := n.populatePool(); err != nil {
		return nil, err
	}
	return &n, nil
}

// populatePool takes the given IP range from configuration and
// adds the IP addresses to the pool for allocation
func (n *networkService) populatePool() error {
	t := n.metrics.NewTiming()
	defer t.Send("populate_pool")
	ip := net.ParseIP(n.conf.Network.IP4.StartAddr)
	ip = ip.To4()
	if ip == nil {
		return errors.New("bad start IP provided in config")
	}
	n.ip4Pool[ip.String()] = nil
	for j := ip[3]; int(j) < n.conf.Network.IP4.Range; j++ {
		ip[3]++
		n.ip4Pool[ip.String()] = nil
	}
	return nil
}

// Allocate checks for available ip addresses returns one
// if available
func (n *networkService) Allocate(id []byte) (string, error) {
	t := n.metrics.NewTiming()
	defer t.Send("allocate")
	n.mu.Lock()
	defer n.mu.Unlock()
	for k := range n.ip4Pool {
		if n.ip4Pool[k] == nil {
			n.ip4Pool[k] = id
			n.metrics.Histogram(k, 1)
			return k, nil
		}
	}
	return "", errors.New("no addresses available")
}

// Return
func (n *networkService) Return(ip string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.ip4Pool[ip] = nil
}

// Pool returns the current state of the IP address pool
func (n *networkService) Pool() map[string][]byte {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.ip4Pool
}

// UpdateIPState iterates through the pool of IP addresses
// and if found sets it to the given state
func (n *networkService) UpdateIPState(ip string, state []byte) error {
	t := n.metrics.NewTiming()
	defer t.Send("update_ip_state")
	n.mu.Lock()
	defer n.mu.Unlock()
	for k := range n.ip4Pool {
		if k == ip {
			n.ip4Pool[k] = state
			return nil
		}
	}
	return errors.New("unknown ip")
}
