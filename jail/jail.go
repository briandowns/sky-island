package jail

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/filesystem"
	"github.com/briandowns/sky-island/utils"
	"gopkg.in/alexcesaro/statsd.v2"
)

const (
	sysDownloadURL = "http://ftp.freebsd.org/pub/FreeBSD/releases/amd64/amd64/%s/%s"
	goDownloadURL  = "https://redirector.gvt1.com/edgedl/go/go%s.freebsd-amd64.tar.gz"
)

const rcConf = "/etc/rc.conf"

var basePackages = []string{"base.txz", "lib32.txz", "ports.txz"}

// JailServicer defines the behavior of the Jail service
type JailServicer interface {
	InitializeSystem() error
	CreateJail(string, bool) error
	RemoveJail(string) error
	KillJail(int) error
	JailDetails(int) (*JLS, error)
}

// jailService holds the state of the service
type jailService struct {
	logger    *logrus.Logger
	conf      *config.Config
	hc        *http.Client
	metrics   *statsd.Client
	fsService filesystem.FSServicer
	wrapper   utils.Wrapper
}

// NewJailService creates a new value of type jailService pointer
func NewJailService(conf *config.Config, l *logrus.Logger, m *statsd.Client, w utils.Wrapper) JailServicer {
	return &jailService{
		logger: l,
		conf:   conf,
		hc: &http.Client{
			Timeout: time.Second * 300,
		},
		metrics:   m,
		fsService: filesystem.NewFilesystemService(conf, l, m, w),
		wrapper:   w,
	}
}

// configureJailHostname sets the hostname of the jail
// to the name of the jail
func (j *jailService) configureJailHostname(name string) error {
	rcconf, err := os.Open(j.conf.Jails.BaseJailDir + "/" + name + "/etc/rc.conf")
	if err != nil {
		return err
	}
	defer rcconf.Close()
	rcconf.Write([]byte("hostname=" + name))
	return nil
}

// InitializeSystem is run to make sure the systsem that will be running
// sky-island has all of the necessary features in place and configured
func (j *jailService) InitializeSystem() error {
	t := j.metrics.NewTiming()
	defer t.Send("initialize_system")
	j.logger.Info("creating ZFS dataset")
	if err := j.fsService.CreateDataset(); err != nil {
		return err
	}
	j.logger.Info("downloading base system")
	if err := j.downloadBaseSystem(); err != nil {
		return err
	}
	j.logger.Info("extracting packages into base jail")
	if err := j.extractBasePkgs(); err != nil {
		return err
	}
	j.logger.Info("updating base jail")
	if err := j.updateBaseJail(); err != nil {
		return err
	}
	j.logger.Info("setting base jail config")
	if err := j.setBaseJailConf(); err != nil {
		return err
	}
	j.logger.Info("installing Go")
	if err := j.installGo(); err != nil {
		return err
	}
	j.logger.Info("creating base jail snapshot")
	if err := j.fsService.CreateSnapshot(); err != nil {
		return err
	}
	j.logger.Info("creating build jail")
	if err := j.CreateJail("build", false); err != nil {
		return err
	}
	j.logger.Info("creating monitoring jail")
	return j.buildMonitoringJail()
}

// CreateJail creates a jail with a name of the given name and
// sets resource limits
func (j *jailService) CreateJail(name string, sl bool) error {
	t := j.metrics.NewTiming()
	defer t.Send("create_jail_time")
	if err := j.fsService.CloneBaseToJail(name); err != nil {
		return err
	}
	f, err := os.Create(j.conf.Jails.BaseJailDir + "/" + name + rcConf)
	if err != nil {
		return err
	}
	defer f.Close()
	if sl {
		if err := j.applyResourceLimits(); err != nil {
			return err
		}
	}
	f.Write([]byte(fmt.Sprintf(`hostname="%s"`, name)))
	j.metrics.Histogram("created", 1)
	return nil
}

// applyResourceLimits
func (j *jailService) applyResourceLimits() error {
	return nil
}

// RemoveJail removes the jail with the given name
func (j *jailService) RemoveJail(name string) error {
	t := j.metrics.NewTiming()
	defer t.Send("remove_jail_time")
	if err := j.fsService.RemoveDataset(name); err != nil {
		j.logger.Error(err)
		return err
	}
	j.metrics.Histogram("removed", 1)
	return nil
}

// KillJail stops a running jail
func (j *jailService) KillJail(id int) error {
	t := j.metrics.NewTiming()
	defer t.Send("kill_jail_time")
	_, err := exec.Command("jail", "-r", strconv.Itoa(id)).Output()
	if err != nil {
		return err
	}
	j.metrics.Histogram("killed", 1)
	return nil
}

// JailDetails runs the system jls command and returns either the
// output or an error
func (j *jailService) JailDetails(id int) (*JLS, error) {
	jails, err := JLSRun(utils.Wrap{})
	if err != nil {
		return nil, err
	}
	for _, jail := range jails {
		if id == jail.JID {
			return jail, nil
		}
	}
	return nil, fmt.Errorf("jail %d not found", id)
}

// startJail starts the given jail
func (j *jailService) startJail(name string) error {
	return nil
}

// devFSMounted determines if the given jail has devfs mounted
func (j *jailService) devFSMounted(name string) bool {
	return false
}

// unmountDevFS
func (j *jailService) unmountDevFS(name string) error {
	return nil
}
