package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestStatsHandler
func TestStatsHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/v1/admin/api-stats", nil)
	if err != nil {
		t.Error(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testHandler.statsHandler())

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
	}
}
