package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/briandowns/sky-island/jail"
)

// statsHandler handles API stats processing requests
func (h *handler) statsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ren.JSON(w, http.StatusOK, h.statsMW.Data())
	}
}

// networkHandler handles requests for the ip service
func (h *handler) networkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vals := r.URL.Query()
		p, ok := vals["state"]
		if !ok {
			h.ren.JSON(w, http.StatusOK, h.networksvc.Pool())
			return
		}
		pool := h.networksvc.Pool()
		if p[0] == "available" {
			var available []string
			for k := range pool {
				if pool[k] == 0 {
					available = append(available, k)
				}
			}
			h.ren.JSON(w, http.StatusOK, map[string][]string{"available": available})
			return
		}
		if p[0] == "unavailable" {
			var unavailable []string
			for k := range pool {
				if pool[k] == 1 {
					unavailable = append(unavailable, k)
				}
			}
			h.ren.JSON(w, http.StatusOK, map[string][]string{"unavailable": unavailable})
			return
		}
		h.ren.JSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "unrecognized IP state"})
	}
}

// ipStateUpdateRequest contains a fields seen
// in a ip state update request
type ipStateUpdateRequest struct {
	IP    string `json:"ip"`
	State byte   `json:"state"`
}

// updateIPStateHandler receives a request to update the state of an IP address
func (h *handler) updateIPStateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			h.logger.Info(err.Error())
			h.ren.JSON(w, http.StatusInternalServerError, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		var req ipStateUpdateRequest
		if err := json.Unmarshal(b, &req); err != nil {
			h.logger.Info(err.Error())
			h.ren.JSON(w, http.StatusBadRequest, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		ip := net.ParseIP(req.IP)
		ip = ip.To4()
		if ip == nil {
			h.ren.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid IP4 address"})
			return
		}
		if !jail.ValidIPState(req.State) {
			h.ren.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid state"})
			return
		}
		if err := h.networksvc.UpdateIPState(ip.String(), req.State); err != nil {
			h.ren.JSON(w, http.StatusOK, map[string]string{"error": http.StatusText(http.StatusInternalServerError)})
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
