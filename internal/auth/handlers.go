package auth

import (
	"encoding/json"
	"net/http"

	httpmw "github.com/mbuitragoc/panini-api/internal/http/middleware"
)

// Handlers groups the HTTP handlers for the auth domain.
type Handlers struct {
	svc *Service
}

// NewHandlers creates a new auth Handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// PostAuthApple handles POST /auth/apple.
// It validates the Apple identity token and returns a JWT.
func (h *Handlers) PostAuthApple(w http.ResponseWriter, r *http.Request) {
	var req AppleTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.AuthenticateApple(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// PostUsersMe handles POST /v1/users/me — upserts the authenticated user's profile.
func (h *Handlers) PostUsersMe(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpsertUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.svc.repo.UpsertProfile(r.Context(), userID, req.Username, req.Handle)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// GetUsersMe handles GET /v1/users/me — returns the authenticated user's profile.
func (h *Handlers) GetUsersMe(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.repo.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// GetUsersSearch handles GET /v1/users/search?handle= — searches users by handle substring.
func (h *Handlers) GetUsersSearch(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		writeError(w, http.StatusBadRequest, "handle query param required")
		return
	}
	users, err := h.svc.SearchByHandle(r.Context(), handle)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if users == nil {
		users = []User{}
	}
	writeJSON(w, http.StatusOK, users)
}

// PostDeviceToken handles POST /v1/users/me/device-token — stores an APNs device token.
func (h *Handlers) PostDeviceToken(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req DeviceTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.repo.SaveDeviceToken(r.Context(), userID, req.DeviceToken); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeJSON encodes v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
