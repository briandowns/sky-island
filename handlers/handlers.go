package handlers

import (
	"net/http"

	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/filesystem"
	"github.com/briandowns/sky-island/jail"
	"github.com/briandowns/sky-island/utils"
	gklog "github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/thoas/stats"
	"github.com/unrolled/render"
	statsd "gopkg.in/alexcesaro/statsd.v2"
)

// Params
type Params struct {
	Conf    *config.Config
	Logger  gklog.Logger
	StatsMW *stats.Stats
	Metrics *statsd.Client
}

// handler
type handler struct {
	ren        *render.Render
	conf       *config.Config
	logger     gklog.Logger
	statsMW    *stats.Stats
	metrics    *statsd.Client
	rsvc       jail.RepoServicer
	networksvc jail.NetworkServicer
	jsvc       jail.JailServicer
	fssvc      filesystem.FSServicer
}

// AddHandlers builds all endpoints to be passed into the
func AddHandlers(router *mux.Router, p *Params) error {
	p.Logger.Log("msg", "initializing route handlers")
	networksvc, err := jail.NewNetworkService(p.Conf, p.Logger, p.Metrics.Clone(statsd.Prefix("network")))
	if err != nil {
		return err
	}
	h := &handler{
		ren:        render.New(),
		conf:       p.Conf,
		logger:     p.Logger,
		statsMW:    p.StatsMW,
		metrics:    p.Metrics,
		rsvc:       jail.NewRepoService(p.Conf, p.Logger, p.Metrics.Clone(statsd.Prefix("repo"))),
		networksvc: networksvc,
		jsvc:       jail.NewJailService(p.Conf, p.Logger, p.Metrics.Clone(statsd.Prefix("jail")), utils.Wrap{}),
		fssvc:      filesystem.NewFilesystemService(p.Conf, p.Logger, p.Metrics.Clone(statsd.Prefix("filesystem")), utils.Wrap{}),
	}
	router.HandleFunc("/healthcheck", h.healthcheckHandler()).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/function", h.functionRunHandler()).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/admin/api-stats", h.statsHandler()).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/admin/jails", h.jailsRunningHandler()).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/admin/jail/{id}", h.jailDetailsHandler()).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/admin/jail/{id}", h.killJailHandler()).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/admin/jails", h.killAllJailsHandler()).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/admin/network/ips", h.networkHandler()).Queries("state", "{state}").Methods(http.MethodGet)
	router.HandleFunc("/api/v1/admin/network/ip", h.updateIPStateHandler()).Methods(http.MethodPut)
	return nil
}

// healthcheckHandler handles all healthcheck requests
func (h *handler) healthcheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
}
