package filesystem

import (
	"fmt"

	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/utils"
	gklog "github.com/go-kit/kit/log"
	"gopkg.in/alexcesaro/statsd.v2"
)

// FSServicer defines the behavior of the filesystem service
type FSServicer interface {
	CreateBaseJailDataset() error
	CloneBaseToJail(string) error
	CreateDataset() error
	CreateSnapshot() error
	RemoveDataset(string) error
}

// fsService
type fsService struct {
	logger  gklog.Logger
	conf    *config.Config
	wrapper utils.Wrapper
	metrics *statsd.Client
}

// NewFilesystemService creates a new value of type FileSystemService which provides the dependencies
// to the service methods
func NewFilesystemService(conf *config.Config, l gklog.Logger, metrics *statsd.Client, w utils.Wrapper) FSServicer {
	return &fsService{
		logger:  l,
		conf:    conf,
		wrapper: w,
		metrics: metrics,
	}
}

// CreateBaseJailDataset creates a Dataset and mounts it for the base jail
func (f *fsService) CreateBaseJailDataset() error {
	t := f.metrics.NewTiming()
	defer t.Send("dataset_create")
	_, err := f.wrapper.Output("zfs", "create", "-o", "mountpoint="+f.conf.Jails.BaseJailDir, f.conf.Filesystem.ZFSDataset, "")
	return err
}

// CloneBaseToJail does a ZFS clone from the base jail to the new jail
func (f *fsService) CloneBaseToJail(jname string) error {
	t := f.metrics.NewTiming()
	defer t.Send("dataset_create")
	base := f.conf.Filesystem.ZFSDataset + "/jails/releases/" + f.conf.Release + "@p1"
	dataset := f.conf.Filesystem.ZFSDataset + "/jails/" + jname
	_, err := f.wrapper.Output("zfs", "clone", base, dataset)
	return err
}

// CreateDataset creates a new ZFS Dataset
func (f *fsService) CreateDataset() error {
	dataset := f.conf.Filesystem.ZFSDataset + "/jails/releases/" + f.conf.Release
	out, err := f.wrapper.CombinedOutput("zfs", "create", "-p", dataset)
	fmt.Println(string(out))
	return err
}

// CreateSnapshot creates a ZFS snapshot of the base jail
func (f *fsService) CreateSnapshot() error {
	t := f.metrics.NewTiming()
	defer t.Send("snapshot_create")
	_, err := f.wrapper.Output("zfs", "snapshot", f.conf.Filesystem.ZFSDataset+"/jails/releases/"+f.conf.Release+"@p1")
	return err
}

// RemoveDataset removes the Dataset associated with the given id
func (f *fsService) RemoveDataset(id string) error {
	t := f.metrics.NewTiming()
	defer t.Send("dataset_remove")
	_, err := f.wrapper.Output("zfs", "destroy", "-rf", f.conf.Filesystem.ZFSDataset+"/jails/"+id)
	return err
}
