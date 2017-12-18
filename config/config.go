package config

import (
	"encoding/json"
	"io/ioutil"
)

// Filesystem
type Filesystem struct {
	ZFSDataset  string `json:"zfs_dataset"`
	Compression bool   `json:"compression"`
}

// Network
type Network struct {
	IP4 *IP4 `json:"ip4"`
}

// IP4 contains necessary components for network connectivity
type IP4 struct {
	Interface string   `json:"interface"`
	StartAddr string   `json:"start_addr"`
	Mask      string   `json:"mask"`
	Range     int      `json:"range"`
	Gateway   string   `json:"gateway"`
	DNS       []string `json:"dns"`
}

// Jails contains necessary components to setup
// the necessary jails
type Jails struct {
	BaseJailDir            string `json:"base_jail_dir"`
	CacheDefaultExpiration string `json:"cache_default_expiration"`
	CachePurgeAfter        string `json:"cache_purge_after"`
	ChildrenMax            int    `json:"children_max"`
	MonitoringAddr         string `json:"monitoring_addr"`
}

// Config contains the parameters necessary to run sky-island
type Config struct {
	Release       string
	SystemPort    int         `json:"system_port"`
	HTTPPort      int         `json:"http_port"`
	GoVersion     string      `json:"go_version"`
	BaseSysPkgDir string      `json:"base_sys_pkg_dir"`
	Filesystem    *Filesystem `json:"filesystem"`
	Network       *Network    `json:"network"`
	Jails         *Jails      `json:"jails"`
}

// Load prses the given file and creates a new value
// of type Config pointer
func Load(confFile string) (*Config, error) {
	f, err := ioutil.ReadFile(confFile)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(f, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
