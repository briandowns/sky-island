package jail

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/briandowns/sky-island/config"
	"gopkg.in/alexcesaro/statsd.v2"
	"gopkg.in/src-d/go-git.v4"
)

// RepoServicer
type RepoServicer interface {
	CloneRepo(jpath, fname string) error
	RemoveRepo(repo string) error
}

// repoService
type repoService struct {
	logger  *logrus.Logger
	conf    *config.Config
	metrics *statsd.Client
}

// newRepoService
func NewRepoService(conf *config.Config, l *logrus.Logger, metrics *statsd.Client) RepoServicer {
	return &repoService{
		logger:  l,
		conf:    conf,
		metrics: metrics,
	}
}

// CloneRepo clones the given repo into the given path
func (r *repoService) CloneRepo(jpath, fname string) error {
	t := r.metrics.NewTiming()
	defer t.Send("clone")
	_, err := git.PlainClone(jpath+"/"+fname, false, &git.CloneOptions{
		URL: "https://" + fname + ".git",
	})
	if err != nil {
		return err
	}
	r.metrics.Histogram(fname, 1)
	return nil
}

// RemoveRepo removes the given repo from the build jail
func (r *repoService) RemoveRepo(repo string) error {
	t := r.metrics.NewTiming()
	defer t.Send("remove")
	if err := os.RemoveAll(r.conf.Jails.BaseJailDir + "/build/root/go/src/" + repo); err != nil {
		return err
	}
	r.metrics.Histogram(repo, 1)
	return nil
}
