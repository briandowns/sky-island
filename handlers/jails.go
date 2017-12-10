package handlers

import (
	"net/http"
	"strconv"

	"github.com/briandowns/sky-island/jail"
	"github.com/briandowns/sky-island/utils"
	"github.com/gorilla/mux"
)

// jailsRunningHandler
func (h *handler) jailsRunningHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jls, err := jail.JLSRun(utils.Wrap{})
		if err != nil {
			h.logger.Error(err)
			h.ren.JSON(w, http.StatusOK, map[string]string{"error": err.Error()})
			return
		}
		h.ren.JSON(w, http.StatusOK, map[string]interface{}{"jails": jls})
		h.metrics.Histogram("handlers.jail.running", 1)
	}
}

// jailID
func jailID(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars["id"])
}

// jailDetailsHandler
func (h *handler) jailDetailsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jid, err := jailID(r)
		if err != nil {
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		jail, err := h.jsvc.JailDetails(jid)
		if err != nil {
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		h.ren.JSON(w, http.StatusOK, map[string]interface{}{"details": jail})
		h.metrics.Histogram("handlers.jail.details", 1)
	}
}

// killJailHandler
func (h *handler) killJailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jid, err := jailID(r)
		if err != nil {
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if err := h.jsvc.KillJail(jid); err != nil {
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		h.ren.JSON(w, http.StatusOK, map[string]int{"deleted": jid})
		h.metrics.Histogram("handlers.jail.kill", 1)
	}
}

// killAllJailsHandler
func (h *handler) killAllJailsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jails, err := jail.JLSRun(utils.Wrap{})
		if err != nil {
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		for _, j := range jails {
			if err := h.jsvc.KillJail(j.JID); err != nil {
				h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			}
		}
		w.WriteHeader(http.StatusOK)
		h.metrics.Histogram("handlers.jail.killall", 1)
	}
}
