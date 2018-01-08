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

const apiPrefix = "/api/v1"

// Params contains the necessary dependencies
// for the handler type and handlers derived
type Params struct {
	Conf    *config.Config
	Logger  gklog.Logger
	StatsMW *stats.Stats
	Metrics *statsd.Client
}

// handler contains the state of the api system
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
	binCache   *jail.BinaryCache
}

// AddHandlers builds all endpoints to be passed into the
func AddHandlers(p *Params) (*mux.Router, error) {
	p.Logger.Log("msg", "initializing route handlers")
	networksvc, err := jail.NewNetworkService(p.Conf, p.Logger, p.Metrics.Clone(statsd.Prefix("network")))
	if err != nil {
		return nil, err
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
		binCache:   jail.NewBinaryCache(),
	}
	router := mux.NewRouter()
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	router.HandleFunc("/healthcheck", h.healthcheckHandler()).Methods(http.MethodGet)

	fr := router.PathPrefix(apiPrefix).Subrouter()
	fr.Path("/function").HandlerFunc(h.functionRunHandler()).Methods(http.MethodPost)

	ar := router.PathPrefix(apiPrefix).Subrouter()
	ar.Path("/admin/api-stats").HandlerFunc(h.auth(h.statsHandler())).Methods(http.MethodGet)
	ar.Path("/admin/jails").HandlerFunc(h.auth(h.jailsRunningHandler())).Methods(http.MethodGet)
	ar.Path("/admin/jail/{id}").HandlerFunc(h.auth(h.jailDetailsHandler())).Methods(http.MethodGet)
	ar.Path("/admin/jail/{id}").HandlerFunc(h.auth(h.killJailHandler())).Methods(http.MethodDelete)
	ar.Path("/admin/jails").HandlerFunc(h.auth(h.killAllJailsHandler())).Methods(http.MethodDelete)
	ar.Path("/admin/network/ips").HandlerFunc(h.auth(h.networkHandler())).Methods(http.MethodGet)
	ar.Path("/admin/network/ips").HandlerFunc(h.auth(h.networkHandler())).Queries("state", "{state}").Methods(http.MethodGet)
	ar.Path("/admin/network/ip").HandlerFunc(h.auth(h.updateIPStateHandler())).Methods(http.MethodPut)
	return router, nil
}

// healthcheckHandler handles all healthcheck requests
func (h *handler) healthcheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
}

// auth checks to see if the configured header and token are provided
// in the request
func (h *handler) auth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var hdr string
		if hdr = r.Header.Get(h.conf.AdminTokenHeader); hdr != h.conf.AdminAPIToken {
			h.logger.Log("error", "unauthorized request received")
			h.ren.JSON(w, http.StatusForbidden, map[string]string{"error": http.StatusText(http.StatusForbidden)})
			return
		}
		fn(w, r)
	}
}
