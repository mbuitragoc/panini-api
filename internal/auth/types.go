// Package auth handles Apple Sign-In, JWT issuance, and user identity.
package auth

import (
	"time"

	"github.com/google/uuid"
)

// AppleTokenRequest is the payload sent by the client to authenticate via Apple Sign-In.
type AppleTokenRequest struct {
	IdentityToken string `json:"identityToken"`
	Nonce         string `json:"nonce"`
}

// AuthResponse is returned after a successful authentication.
type AuthResponse struct {
	JWT       string `json:"jwt"`
	IsNewUser bool   `json:"isNewUser"`
	UserID    string `json:"userID"`
}

// User represents an authenticated application user.
type User struct {
	ID        uuid.UUID `json:"id"`
	AppleSub  string    `json:"appleSub"`
	Username  string    `json:"username"`
	Handle    string    `json:"handle"`
	CreatedAt time.Time `json:"createdAt"`
}

// UpsertUserRequest carries optional profile fields the client may set.
type UpsertUserRequest struct {
	Username string `json:"username"`
	Handle   string `json:"handle"`
}

// DeviceTokenRequest carries an APNs device token for push notifications.
type DeviceTokenRequest struct {
	DeviceToken string `json:"deviceToken"`
}
