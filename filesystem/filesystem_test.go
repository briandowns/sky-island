package filesystem

import (
	"testing"

	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/utils"
	gklog "github.com/go-kit/kit/log"
	"gopkg.in/alexcesaro/statsd.v2"
)

var testConf = &config.Config{
	Filesystem: &config.Filesystem{
		ZFSDataset:  "test/dataset",
		Compression: false,
	},
	Jails: &config.Jails{
		BaseJailDir: "test/dir",
	},
}

// TestNewFilesystemService
func TestNewFilesystemService(t *testing.T) {
	fsSvc := NewFilesystemService(testConf, gklog.NewNopLogger(), &statsd.Client{}, utils.NoOpWrapper{})
	if fsSvc == nil {
		t.Error("expected not nil filesystem service")
	}
}

// TestCreateBaseJailDataset
func TestCreateBaseJailDataset(t *testing.T) {
	fsSvc := NewFilesystemService(testConf, gklog.NewNopLogger(), &statsd.Client{}, utils.NoOpWrapper{})
	if fsSvc == nil {
		t.Error("expected not nil filesystem service")
	}
	if err := fsSvc.CreateBaseJailDataset(); err != nil {
		t.Error(err)
	}
}

func TestCloneBaseToJail(t *testing.T) {
	fsSvc := NewFilesystemService(testConf, gklog.NewNopLogger(), &statsd.Client{}, utils.NoOpWrapper{})
	if fsSvc == nil {
		t.Error("expected not nil filesystem service")
	}
	if err := fsSvc.CloneBaseToJail("test-jail-name"); err != nil {
		t.Error(err)
	}
}

// TestCreateDataset
func TestCreateDataset(t *testing.T) {
	fsSvc := NewFilesystemService(testConf, gklog.NewNopLogger(), &statsd.Client{}, utils.NoOpWrapper{})
	if fsSvc == nil {
		t.Error("expected not nil filesystem service")
	}
	if err := fsSvc.CreateDataset(); err != nil {
		t.Error(err)
	}
}

// TestCreateSnapshot
func TestCreateSnapshot(t *testing.T) {
	fsSvc := NewFilesystemService(testConf, gklog.NewNopLogger(), &statsd.Client{}, utils.NoOpWrapper{})
	if fsSvc == nil {
		t.Error("expected not nil filesystem service")
	}
	if err := fsSvc.CreateSnapshot(); err != nil {
		t.Error(err)
	}
}

// TestRemoveDataset
func TestRemoveDataset(t *testing.T) {
	fsSvc := NewFilesystemService(testConf, gklog.NewNopLogger(), &statsd.Client{}, utils.NoOpWrapper{})
	if fsSvc == nil {
		t.Error("expected not nil filesystem service")
	}
	if err := fsSvc.RemoveDataset("test-Dataset"); err != nil {
		t.Error(err)
	}
}
