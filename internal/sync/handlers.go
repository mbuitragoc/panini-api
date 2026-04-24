package sync

import (
	"encoding/json"
	"net/http"

	httpmw "github.com/mbuitragoc/panini-api/internal/http/middleware"
)

// Handlers groups HTTP handlers for the sync domain.
type Handlers struct {
	svc *Service
}

// NewHandlers creates a new sync Handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// GetSync handles GET /v1/sync.
func (h *Handlers) GetSync(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.svc.Sync(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
