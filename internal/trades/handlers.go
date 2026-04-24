package trades

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httpmw "github.com/mbuitragoc/panini-api/internal/http/middleware"
)

// Handlers groups HTTP handlers for the trades domain.
type Handlers struct {
	svc *Service
}

// NewHandlers creates a new trades Handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// GetTrades handles GET /v1/trades.
func (h *Handlers) GetTrades(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	trades, err := h.svc.ListForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, trades)
}

// PostTrade handles POST /v1/trades.
func (h *Handlers) PostTrade(w http.ResponseWriter, r *http.Request) {
	userID := httpmw.UserIDFromContext(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	trade, err := h.svc.Create(r.Context(), userID, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, trade)
}

// PutTradeAccept handles PUT /v1/trades/{id}/accept.
func (h *Handlers) PutTradeAccept(w http.ResponseWriter, r *http.Request) {
	h.applyEvent(w, r, EventAccept)
}

// PutTradeDecline handles PUT /v1/trades/{id}/decline.
func (h *Handlers) PutTradeDecline(w http.ResponseWriter, r *http.Request) {
	h.applyEvent(w, r, EventDecline)
}

// PutTradeConfirm handles PUT /v1/trades/{id}/confirm.
func (h *Handlers) PutTradeConfirm(w http.ResponseWriter, r *http.Request) {
	h.applyEvent(w, r, EventConfirm)
}

// applyEvent is a shared helper that applies a trade event from the authenticated user.
func (h *Handlers) applyEvent(w http.ResponseWriter, r *http.Request, event TradeEvent) {
	rawUserID := httpmw.UserIDFromContext(r.Context())
	if rawUserID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	actorID, err := uuid.Parse(rawUserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid user id")
		return
	}

	tradeID := chi.URLParam(r, "id")
	if tradeID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	trade, err := h.svc.Apply(r.Context(), tradeID, actorID, event)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, trade)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
