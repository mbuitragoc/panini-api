package collections

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpmw "github.com/mbuitragoc/panini-api/internal/http/middleware"
)

// Handlers groups HTTP handlers for the collections domain.
type Handlers struct {
	svc *Service
}

// NewHandlers creates a new collections Handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// GetCollections handles GET /v1/collections.
func (h *Handlers) GetCollections(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.svc.ListForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// PutCollection handles PUT /v1/collections/{stickerID}.
func (h *Handlers) PutCollection(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	stickerID := chi.URLParam(r, "stickerID")
	if stickerID == "" {
		writeError(w, http.StatusBadRequest, "stickerID is required")
		return
	}

	var req UpsertCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.svc.Upsert(r.Context(), userID, stickerID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
