package config

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

// System
type System struct {
	Port          int
	BaseSysPkgDir string
	GoVersion     string
	LogFile       string
}

// HTTP
type HTTP struct {
	Port         int
	FrontendPath string
}

// Filesystem
type Filesystem struct {
	ZFSDataset  string
	Compression bool
}

// IP4 contains necessary components for network connectivity
type IP4 struct {
	Interface string
	StartAddr string
	Mask      string
	Range     int
	Gateway   string
	DNS       []string
}

// Jails contains necessary components to setup
// the necessary jails
type Jails struct {
	BaseJailDir            string
	CacheDefaultExpiration string
	CachePurgeAfter        string
	ChildrenMax            int
	MonitoringAddr         string
}

// Config contains the parameters necessary to run sky-island
type Config struct {
	Release    string
	HTTP       HTTP
	Filesystem Filesystem
	System     System
	Jails      Jails
	IP4        IP4
}

// Load prses the given file and creates a new value
// of type Config pointer
func Load(confFile string) (*Config, error) {
	f, err := ioutil.ReadFile(confFile)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := toml.Unmarshal(f, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
