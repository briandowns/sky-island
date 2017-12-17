package jail

import (
	"os"
	"testing"

	gklog "github.com/go-kit/kit/log"
	"gopkg.in/alexcesaro/statsd.v2"
)

// removeTempRepos removes the repos cloned during testing
func removeTempRepos() {
	os.RemoveAll("/tmp/github.com")
}

// TestCloneRepo verifies that a repo is successfully cloned
// from the given URL
func TestCloneRepo(t *testing.T) {
	defer removeTempRepos()
	rs := NewRepoService(testConf, gklog.NewNopLogger(), &statsd.Client{})
	if err := rs.CloneRepo("/tmp", "github.com/briandowns/smile"); err != nil {
		t.Error(err)
	}
}

// TestCloneRepo_Failure verifies that an error is returned
// when trying to clone a repo from a bad URL
func TestCloneRepo_Failure(t *testing.T) {
	defer removeTempRepos()
	rs := NewRepoService(testConf, gklog.NewNopLogger(), &statsd.Client{})
	if err := rs.CloneRepo("/tmp", "github.com/briandown/smile"); err == nil {
		t.Error("expected error but received none")
	}
}
