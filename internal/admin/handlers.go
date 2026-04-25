package admin

import (
	"encoding/json"
	"net/http"

	httpmw "github.com/mbuitragoc/panini-api/internal/http/middleware"
)

// Handlers groups HTTP handlers for admin operations.
type Handlers struct {
	repo *Repo
}

// NewHandlers creates a new admin Handlers.
func NewHandlers(repo *Repo) *Handlers {
	return &Handlers{repo: repo}
}

type missingRatingsRequest struct {
	StickerIDs []string `json:"sticker_ids"`
}

// PostMissingRatings handles POST /v1/admin/missing-ratings.
// Accepts a list of sticker IDs the client could not match to rating data.
// Returns 204 No Content — fire-and-forget from the client's perspective.
func (h *Handlers) PostMissingRatings(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req missingRatingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.StickerIDs) == 0 {
		w.WriteHeader(http.StatusNoContent) // malformed or empty — silently accept
		return
	}

	// Cap to 100 IDs per request to prevent abuse.
	if len(req.StickerIDs) > 100 {
		req.StickerIDs = req.StickerIDs[:100]
	}

	_ = h.repo.ReportMissingRatings(r.Context(), userID, req.StickerIDs)
	w.WriteHeader(http.StatusNoContent)
}
