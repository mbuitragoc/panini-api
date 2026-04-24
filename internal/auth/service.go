package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service implements authentication business logic.
type Service struct {
	repo      *Repo
	jwtSecret []byte
	apple     *appleValidator
}

// NewService creates a new auth Service.
func NewService(repo *Repo, jwtSecret, appleBundleID string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
		apple:     newAppleValidator(appleBundleID),
	}
}

// AuthenticateApple validates the Apple identity token, upserts the user, and issues a JWT.
func (s *Service) AuthenticateApple(ctx context.Context, req AppleTokenRequest) (*AuthResponse, error) {
	if req.IdentityToken == "" {
		return nil, fmt.Errorf("auth: identity token must not be empty")
	}

	appleSub, err := s.apple.ValidateToken(ctx, req.IdentityToken)
	if err != nil {
		return nil, fmt.Errorf("auth: invalid apple token: %w", err)
	}

	existing, err := s.repo.GetByAppleSub(ctx, appleSub)
	if err != nil {
		return nil, fmt.Errorf("auth: look up apple sub: %w", err)
	}

	isNewUser := existing == nil

	var user *User
	if isNewUser {
		user, err = s.repo.Create(ctx, appleSub)
		if err != nil {
			return nil, fmt.Errorf("auth: create user: %w", err)
		}
	} else {
		user = existing
	}

	token, err := s.issueJWT(user)
	if err != nil {
		return nil, fmt.Errorf("auth: issue jwt: %w", err)
	}

	return &AuthResponse{
		JWT:       token,
		IsNewUser: isNewUser,
		UserID:    user.ID.String(),
	}, nil
}

// SearchByHandle returns users whose handle matches the given substring.
func (s *Service) SearchByHandle(ctx context.Context, handle string) ([]User, error) {
	return s.repo.SearchByHandle(ctx, handle)
}

// issueJWT creates a signed JWT for the given user valid for 90 days.
func (s *Service) issueJWT(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(90 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}
