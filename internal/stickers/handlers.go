package stickers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handlers groups HTTP handlers for the stickers catalog.
type Handlers struct {
	repo *Repo
}

// NewHandlers creates a new stickers Handlers.
func NewHandlers(repo *Repo) *Handlers {
	return &Handlers{repo: repo}
}

// GetStickers handles GET /v1/stickers.
func (h *Handlers) GetStickers(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// GetSticker handles GET /v1/stickers/{id}.
func (h *Handlers) GetSticker(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	sticker, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if sticker == nil {
		writeError(w, http.StatusNotFound, "sticker not found")
		return
	}

	writeJSON(w, http.StatusOK, sticker)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
