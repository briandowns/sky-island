package jail

import (
	"os"
	"sync"

	"github.com/briandowns/sky-island/config"
	gklog "github.com/go-kit/kit/log"
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
	logger  gklog.Logger
	conf    *config.Config
	metrics *statsd.Client
}

// newRepoService
func NewRepoService(conf *config.Config, l gklog.Logger, metrics *statsd.Client) RepoServicer {
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

// BinaryCache holds the path to compiled binaries
type BinaryCache struct {
	mu    sync.RWMutex
	cache map[string]string
}

// NewBinaryCache creates a new value of type
// BinaryCache pointer. This stores a URL as
// the key and a path to the compiled binary
// as the value
func NewBinaryCache() *BinaryCache {
	return &BinaryCache{
		mu:    sync.RWMutex{},
		cache: make(map[string]string),
	}
}

// Get takes a key as an argument and gets the associated
// value if it exists
func (b *BinaryCache) Get(k string) string {
	b.mu.RLock()
	p, ok := b.cache[k]
	b.mu.RUnlock()
	if !ok {
		return ""
	}
	return p
}

// Set takes a key and a value and sets them in the cache
func (b *BinaryCache) Set(k, v string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cache[k] = v
}
