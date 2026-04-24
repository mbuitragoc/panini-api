package friends

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpmw "github.com/mbuitragoc/panini-api/internal/http/middleware"
)

// Handlers groups HTTP handlers for the friends domain.
type Handlers struct {
	svc *Service
}

// NewHandlers creates a new friends Handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// GetFriends handles GET /v1/friends.
func (h *Handlers) GetFriends(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	friends, err := h.svc.ListFriends(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, friends)
}

// PostFriendRequest handles POST /v1/friends/requests.
func (h *Handlers) PostFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req SendFriendRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	friendship, err := h.svc.SendRequest(r.Context(), userID, req.FriendID.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, friendship)
}

// PutFriendRequest handles PUT /v1/friends/requests/{id}.
func (h *Handlers) PutFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req RespondFriendRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	friendship, err := h.svc.RespondToRequest(r.Context(), requestID, userID, req.Accept)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, friendship)
}

// GetFriendCollection handles GET /v1/friends/{id}/collection.
func (h *Handlers) GetFriendCollection(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "not implemented"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
