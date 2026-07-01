package server

import (
	"errors"
	"net/http"
)

func (h *handlers) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handlers) readyz(w http.ResponseWriter, _ *http.Request) {
	if _, err := h.auth.load(); err != nil && !errors.Is(err, errAuthNotFound) {
		httpError(w, http.StatusServiceUnavailable, "auth state unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
