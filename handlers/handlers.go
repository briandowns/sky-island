package handlers

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/filesystem"
	"github.com/briandowns/sky-island/jail"
	"github.com/briandowns/sky-island/utils"
	"github.com/gorilla/mux"
	"github.com/thoas/stats"
	"github.com/unrolled/render"
	statsd "gopkg.in/alexcesaro/statsd.v2"
)

// Params
type Params struct {
	Conf    *config.Config
	Logger  *logrus.Logger
	StatsMW *stats.Stats
	Metrics *statsd.Client
}

// handler
type handler struct {
	ren     *render.Render
	conf    *config.Config
	logger  *logrus.Logger
	statsMW *stats.Stats
	metrics *statsd.Client
	rsvc    jail.RepoServicer
	ipsvc   jail.IPServicer
	jsvc    jail.JailServicer
	fssvc   filesystem.FSServicer
}

// AddHandlers builds all endpoints to be passed into the
func AddHandlers(router *mux.Router, params *Params) error {
	params.Logger.Info("initializing route handlers")
	ipsvc, err := jail.NewipService(params.Conf, params.Logger, params.Metrics.Clone(statsd.Prefix("ip")))
	if err != nil {
		return err
	}
	h := &handler{
		ren:     render.New(),
		conf:    params.Conf,
		logger:  params.Logger,
		statsMW: params.StatsMW,
		metrics: params.Metrics,
		rsvc:    jail.NewRepoService(params.Conf, params.Logger, params.Metrics.Clone(statsd.Prefix("repo"))),
		ipsvc:   ipsvc,
		jsvc:    jail.NewJailService(params.Conf, params.Logger, params.Metrics.Clone(statsd.Prefix("jail")), utils.Wrap{}),
		fssvc:   filesystem.NewFilesystemService(params.Conf, params.Logger, params.Metrics.Clone(statsd.Prefix("filesystem")), utils.Wrap{}),
	}
	router.HandleFunc("/healthcheck", h.healthcheckHandler()).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/function", h.functionRunHandler()).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/admin/api-stats", h.statsHandler()).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/admin/jails", h.jailsRunningHandler()).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/admin/jail/{id}", h.jailDetailsHandler()).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/admin/jail/{id}", h.killJailHandler()).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/admin/jails", h.killAllJailsHandler()).Methods(http.MethodDelete)
	router.HandleFunc("/api/v1/admin/ips", h.ipsHandler()).Queries("state", "{state}").Methods(http.MethodGet)
	router.HandleFunc("/api/v1/admin/ips", h.updateIPStateHandler()).Methods(http.MethodPut)
	return nil
}

// healthcheckHandler handles all healthcheck requests
func (h *handler) healthcheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
}
