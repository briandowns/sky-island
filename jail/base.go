package jail

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/briandowns/sky-island/utils"
	"github.com/mholt/archiver"
	"golang.org/x/sync/errgroup"
)

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

	return cmd.Wait()
}

// setupResolveConf takes the DNS servers from configuration and
// adds them to the release jail's /etc/resolv.conf but if those
// settings aren't present in configuration, it will copy from
// the host system to thet release jail
func (j *jailService) setupResolvConf() error {
	if j.conf.Network.IP4.DNS != nil {
		out, err := os.Create(j.conf.Jails.BaseJailDir + "/releases/" + j.conf.Release + "/etc/resolv.conf")
		if err != nil {
			return err
		}
		defer out.Close()
		for _, i := range j.conf.Network.IP4.DNS {
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
	return nil
}

// downloadGo downloads the configured version of Go only if it hasn't
// been downloaded previously
func (j *jailService) downloadGo() error {
	t := j.metrics.NewTiming()
	defer t.Send("go.download_time")
	goTarBall := fmt.Sprintf("/tmp/go%s.freebsd-amd64.tar.gz", j.conf.GoVersion)
	if utils.Exists(goTarBall) {
		return nil
	}
	out, err := os.Create(goTarBall)
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf(goDownloadURL, j.conf.GoVersion)
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
	src := fmt.Sprintf("/tmp/go%s.freebsd-amd64.tar.gz", j.conf.GoVersion)
	dst := fmt.Sprintf("%s/releases/%s/usr/local", j.conf.Jails.BaseJailDir, j.conf.Release)
	if err := archiver.TarGz.Open(src, dst); err != nil {
		return err
	}
	if err := j.setupGoEnv(); err != nil {
		return err
	}
	return nil
}
