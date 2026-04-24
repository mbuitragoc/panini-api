package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	appleJWKSURL = "https://appleid.apple.com/auth/keys"
	appleIssuer  = "https://appleid.apple.com"
	jwksTTL      = 24 * time.Hour
)

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleJWKSResponse struct {
	Keys []appleJWK `json:"keys"`
}

// appleValidator validates Apple identity tokens against Apple's public JWKS.
type appleValidator struct {
	bundleID string

	mu        sync.RWMutex
	keys      []appleJWK
	fetchedAt time.Time
}

func newAppleValidator(bundleID string) *appleValidator {
	return &appleValidator{bundleID: bundleID}
}

// ValidateToken validates an Apple identity token and returns the Apple user subject (sub claim).
func (v *appleValidator) ValidateToken(ctx context.Context, rawToken string) (string, error) {
	token, err := jwt.Parse(rawToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("apple: unexpected signing method: %v", t.Header["alg"])
		}
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("apple: missing kid in token header")
		}
		return v.getPublicKey(ctx, kid)
	},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithIssuer(appleIssuer),
		jwt.WithAudience(v.bundleID),
	)
	if err != nil {
		return "", fmt.Errorf("apple: validate token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("apple: invalid claims type")
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return "", fmt.Errorf("apple: missing sub claim")
	}

	return sub, nil
}

func (v *appleValidator) getPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	keys, err := v.getKeys(ctx)
	if err != nil {
		return nil, err
	}

	for _, k := range keys {
		if k.Kid == kid {
			return jwkToRSAPublicKey(k)
		}
	}

	// Key not found — force refresh in case Apple rotated keys.
	v.mu.Lock()
	v.fetchedAt = time.Time{}
	v.mu.Unlock()

	keys, err = v.getKeys(ctx)
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		if k.Kid == kid {
			return jwkToRSAPublicKey(k)
		}
	}

	return nil, fmt.Errorf("apple: no key found for kid %q", kid)
}

func (v *appleValidator) getKeys(ctx context.Context) ([]appleJWK, error) {
	v.mu.RLock()
	if time.Since(v.fetchedAt) < jwksTTL && len(v.keys) > 0 {
		keys := v.keys
		v.mu.RUnlock()
		return keys, nil
	}
	v.mu.RUnlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check under write lock.
	if time.Since(v.fetchedAt) < jwksTTL && len(v.keys) > 0 {
		return v.keys, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appleJWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("apple: build jwks request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apple: fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apple: jwks endpoint returned %d", resp.StatusCode)
	}

	var jwks appleJWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("apple: decode jwks: %w", err)
	}

	v.keys = jwks.Keys
	v.fetchedAt = time.Now()
	return v.keys, nil
}

func jwkToRSAPublicKey(k appleJWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("apple: decode jwk.N: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("apple: decode jwk.E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	var e int
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}
