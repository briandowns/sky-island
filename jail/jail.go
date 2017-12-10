package jail

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/filesystem"
	"github.com/briandowns/sky-island/utils"
	"github.com/mholt/archiver"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alexcesaro/statsd.v2"
)

const (
	sysDownloadURL = "http://ftp.freebsd.org/pub/FreeBSD/releases/amd64/amd64/%s/%s"
	goDownloadURL  = "https://redirector.gvt1.com/edgedl/go/go%s.freebsd-amd64.tar.gz"
)

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

// downloadBaseSystem downloads the base FreeBSD system. It
// will only download the packages if they haven't already
// been downloaded
func (j *jailService) downloadBaseSystem() error {
	t := j.metrics.NewTiming()
	defer t.Send("base.download_packages_time")
	if !utils.Exists("/tmp/" + j.conf.Release) {
		if err := os.Mkdir("/tmp/"+j.conf.Release, os.ModePerm); err != nil {
			return err
		}
	}
	var g errgroup.Group
	for _, p := range basePackages {
		if utils.Exists("/tmp/" + j.conf.Release + "/" + p) {
			continue
		}
		pkg := p
		g.Go(func() error {
			out, err := os.Create("/tmp/" + j.conf.Release + "/" + pkg)
			if err != nil {
				return err
			}
			defer out.Close()
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(sysDownloadURL, j.conf.Release, pkg), nil)
			if err != nil {
				return err
			}
			res, err := j.hc.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			_, err = io.Copy(out, res.Body)
			return err
		})
	}
	return g.Wait()
}

// extractBasePkgs extracts the downlaoded txz files into
// the given path to form the base jail
func (j *jailService) extractBasePkgs() error {
	t := j.metrics.NewTiming()
	defer t.Send("base.extract_packages_time")
	fullPath := j.conf.Jails.BaseJailDir + "/releases/" + j.conf.Release
	for _, p := range basePackages {
		if err := archiver.TarXZ.Open("/tmp/"+j.conf.Release+"/"+p, fullPath); err != nil {
			return err
		}
	}
	return nil
}

// updateBaseJail uses freebsd-update to make sure that
// the base jail is up to date
func (j *jailService) updateBaseJail() error {
	t := j.metrics.NewTiming()
	defer t.Send("base.update_time")
	path := j.conf.Jails.BaseJailDir + "/releases/" + j.conf.Release
	cmd := exec.Command("env", "UNAME_r="+j.conf.Release, "freebsd-update", "-b", "--not-running-from-cron", path, "fetch", "install")
	env := os.Environ()
	env = append(env, "UNAME_r="+j.conf.Release)
	if err := cmd.Start(); err != nil {
		return err
	}
	err := cmd.Wait()
	return err
}

// setupResolveConf takes the DNS servers from configuration and
// adds them to the release jail's /etc/resolv.conf but if those
// settings aren't present in configuration, it will copy from
// the host system to thet release jail
func (j *jailService) setupResolvConf() error {
	if j.conf.IP4.DNS != nil {
		out, err := os.Create(j.conf.Jails.BaseJailDir + "/releases/" + j.conf.Release + "/etc/resolv.conf")
		if err != nil {
			return err
		}
		defer out.Close()
		for _, i := range j.conf.IP4.DNS {
			out.Write([]byte("nameserver " + i + "\n"))
		}
		return nil
	}
	in, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(j.conf.Jails.BaseJailDir + "/releases/" + j.conf.Release + "/etc/resolv.conf")
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// setupLocaltime copies the localtime file from the host
// to the release jail
func (j *jailService) setupLocaltime() error {
	in, err := os.Open("/etc/localtime")
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(j.conf.Jails.BaseJailDir + "/releases/" + j.conf.Release + "/etc/localtime")
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// setBaseJailConf configures the jail to have the same resolv.conf
// and localtime as the host system
func (j *jailService) setBaseJailConf() error {
	if err := j.setupResolvConf(); err != nil {
		return err
	}
	return j.setupLocaltime()
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

// downloadGo downloads the configured version of Go only if it hasn't
// been downloaded previously
func (j *jailService) downloadGo() error {
	t := j.metrics.NewTiming()
	defer t.Send("go.download_time")
	goTarBall := fmt.Sprintf("/tmp/go%s.freebsd-amd64.tar.gz", j.conf.System.GoVersion)
	if utils.Exists(goTarBall) {
		return nil
	}
	out, err := os.Create(goTarBall)
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf(goDownloadURL, j.conf.System.GoVersion)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	res, err := j.hc.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, err = io.Copy(out, res.Body)
	return err
}

// setupGoEnv creates the Go workspace
func (j *jailService) setupGoEnv() error {
	t := j.metrics.NewTiming()
	defer t.Send("go.setup_env_time")
	for _, dir := range []string{"src", "bin", "pkg"} {
		d := fmt.Sprintf("%s/releases/%s/root/go/%s", j.conf.Jails.BaseJailDir, j.conf.Release, dir)
		if err := os.MkdirAll(d, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// installGo installs the configured version of Go
func (j *jailService) installGo() error {
	t := j.metrics.NewTiming()
	defer t.Send("go.install_time")
	if err := j.downloadGo(); err != nil {
		return err
	}
	src := fmt.Sprintf("/tmp/go%s.freebsd-amd64.tar.gz", j.conf.System.GoVersion)
	dst := fmt.Sprintf("%s/releases/%s/usr/local", j.conf.Jails.BaseJailDir, j.conf.Release)
	if err := archiver.TarGz.Open(src, dst); err != nil {
		return err
	}
	if err := j.setupGoEnv(); err != nil {
		return err
	}
	return nil
}

// CreateJail creates a jail with a name of the given name and
// sets resource limits
func (j *jailService) CreateJail(name string, sl bool) error {
	t := j.metrics.NewTiming()
	defer t.Send("create_jail_time")
	if err := j.fsService.CloneBaseToJail(name); err != nil {
		return err
	}
	f, err := os.Create(j.conf.Jails.BaseJailDir + "/" + name + "/etc/rc.conf")
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

// buildMonitoringJail
func (j *jailService) buildMonitoringJail() error {
	if err := j.CreateJail("monitoring", false); err != nil {
		return err
	}
	// update /etc/jail.conf
	// start: jail -c monitoring
	// install monitoring software
	return nil
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

// JLS holds the Go represented output from the
// jls command
type JLS struct {
	Host      string `json:"host"`
	IP4       string `json:"ip4"`
	IP6       string `json:"ip6"`
	JID       int    `json:"jid"`
	Name      string `json:"name"`
	OSRelease string `json:"OSRelease"`
	Path      string `json:"path"`
	Hostname  string `json:"hostname"`
}

// JLSRun runs the jls command to get a slice of the
// running jails
func JLSRun(w utils.Wrapper) ([]*JLS, error) {
	res, err := w.CombinedOutput("jls", "-s")
	if err != nil {
		return nil, errors.New(string(res))
	}
	var jlsData []*JLS
	for _, jd := range bytes.Split(res, []byte("\n")) {
		if len(jd) <= 1 {
			continue
		}
		jail, err := unmarshalJLS(jd)
		if err != nil {
			return nil, err
		}
		jlsData = append(jlsData, jail)
	}
	return jlsData, nil
}

// unmarshal unmarshals the data from a JLS command
// into the given struct
func unmarshalJLS(data []byte) (*JLS, error) {
	kvs := strings.Split(string(data), " ")
	var j JLS
	for _, i := range kvs {
		kv := strings.Split(i, "=")
		switch kv[0] {
		case "host":
			j.Host = kv[1]
		case "ip4":
			j.IP4 = kv[1]
		case "ip6":
			j.IP6 = kv[1]
		case "jid":
			jid, err := strconv.Atoi(kv[1])
			if err != nil {
				return nil, err
			}
			j.JID = jid
		case "name":
			j.Name = kv[1]
		case "osrelease":
			j.OSRelease = kv[1]
		case "path":
			j.Path = kv[1]
		case "host.hostname":
			j.Hostname = kv[1]
		}
	}
	return &j, nil
}
