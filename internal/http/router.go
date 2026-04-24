// Package http wires the Chi router and connects all domain handlers.
package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/mbuitragoc/panini-api/internal/app"
	"github.com/mbuitragoc/panini-api/internal/auth"
	"github.com/mbuitragoc/panini-api/internal/collections"
	"github.com/mbuitragoc/panini-api/internal/friends"
	httpmw "github.com/mbuitragoc/panini-api/internal/http/middleware"
	"github.com/mbuitragoc/panini-api/internal/stickers"
	appsync "github.com/mbuitragoc/panini-api/internal/sync"
	"github.com/mbuitragoc/panini-api/internal/trades"
)

// NewRouter constructs the Chi router, registers all middleware and routes,
// and wires each route to its domain handler.
func NewRouter(a *app.App) http.Handler {
	r := chi.NewRouter()

	// Global middleware stack.
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(120 * time.Second))
	r.Use(chimw.Logger)

	// Domain repositories.
	authRepo := auth.NewRepo(a.DB)
	stickersRepo := stickers.NewRepo(a.DB)
	collectionsRepo := collections.NewRepo(a.DB)
	friendsRepo := friends.NewRepo(a.DB)
	tradesRepo := trades.NewRepo(a.DB)

	// Domain services.
	authSvc := auth.NewService(authRepo, a.Config.JWTSecret, a.Config.AppleClientID)
	collectionsSvc := collections.NewService(collectionsRepo)
	friendsSvc := friends.NewService(friendsRepo)
	tradesSvc := trades.NewService(tradesRepo)
	syncSvc := appsync.NewService(collectionsRepo, tradesRepo, friendsRepo)

	// Domain handlers.
	authH := auth.NewHandlers(authSvc)
	collectionsH := collections.NewHandlers(collectionsSvc)
	friendsH := friends.NewHandlers(friendsSvc)
	tradesH := trades.NewHandlers(tradesSvc)
	syncH := appsync.NewHandlers(syncSvc)

	// Sticker handlers (read-only catalog, no dedicated service layer needed).
	stickersH := stickers.NewHandlers(stickersRepo)

	// Public routes — no authentication required.
	r.Get("/health", handleHealth)
	r.Post("/auth/apple", authH.PostAuthApple)

	// Protected routes — require a valid JWT.
	r.Group(func(r chi.Router) {
		r.Use(httpmw.ValidateJWT(a.Config.JWTSecret))

		// User profile.
		r.Post("/v1/users/me", authH.PostUsersMe)
		r.Get("/v1/users/me", authH.GetUsersMe)
		r.Post("/v1/users/me/device-token", authH.PostDeviceToken)

		// Bulk sync.
		r.Get("/v1/sync", syncH.GetSync)

		// Sticker catalog.
		r.Get("/v1/stickers", stickersH.GetStickers)
		r.Get("/v1/stickers/{id}", stickersH.GetSticker)

		// Collections.
		r.Get("/v1/collections", collectionsH.GetCollections)
		r.Put("/v1/collections/{stickerID}", collectionsH.PutCollection)

		// Friends.
		r.Get("/v1/friends", friendsH.GetFriends)
		r.Post("/v1/friends/requests", friendsH.PostFriendRequest)
		r.Put("/v1/friends/requests/{id}", friendsH.PutFriendRequest)
		r.Get("/v1/friends/{id}/collection", friendsH.GetFriendCollection)

		// Trades.
		r.Get("/v1/trades", tradesH.GetTrades)
		r.Post("/v1/trades", tradesH.PostTrade)
		r.Put("/v1/trades/{id}/accept", tradesH.PutTradeAccept)
		r.Put("/v1/trades/{id}/decline", tradesH.PutTradeDecline)
		r.Put("/v1/trades/{id}/confirm", tradesH.PutTradeConfirm)
	})

	return r
}

// handleHealth returns a simple 200 OK with service status.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
