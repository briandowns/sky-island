package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/mocks"
	"github.com/thoas/stats"
	"github.com/unrolled/render"
)

// mocks for testing
var (
	mockJailSvc    = &mocks.JailServicer{}
	mockRepoSvc    = &mocks.RepoServicer{}
	mockFSSvc      = &mocks.FSServicer{}
	mockNetworkSvc = &mocks.NetworkServicer{}
)

var testConf = &config.Config{
	System: config.System{
		LogFile:       "test-log-file.log",
		GoVersion:     "1.9.2",
		BaseSysPkgDir: "/tmp",
	},
	HTTP: config.HTTP{
		Port:         3280,
		FrontendPath: "handlers/public/html",
	},
	Jails: config.Jails{
		BaseJailDir: "/some/jail/dir",
	},
	Filesystem: config.Filesystem{
		ZFSDataset:  "/zroot/jails",
		Compression: false,
	},
}

var testHandler = &handler{
	conf:       testConf,
	logger:     &logrus.Logger{},
	ren:        render.New(),
	statsMW:    stats.New(),
	jsvc:       mockJailSvc,
	networksvc: mockNetworkSvc,
	fssvc:      mockFSSvc,
}

// TestHealthCheckHandler
func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/healthcheck", nil)
	if err != nil {
		t.Error(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testHandler.healthcheckHandler())
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
	}
}
