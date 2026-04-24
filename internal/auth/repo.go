package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo handles persistence for the auth domain.
type Repo struct {
	db *pgxpool.Pool
}

// NewRepo creates a new auth Repo.
func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

// GetByAppleSub retrieves a user by their Apple subject identifier.
// Returns nil, nil when no user is found.
func (r *Repo) GetByAppleSub(ctx context.Context, appleSub string) (*User, error) {
	const q = `
		SELECT id, apple_sub, COALESCE(username, ''), COALESCE(handle, ''), created_at
		FROM users WHERE apple_sub = $1`

	var u User
	err := r.db.QueryRow(ctx, q, appleSub).
		Scan(&u.ID, &u.AppleSub, &u.Username, &u.Handle, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Create inserts a new user record and returns the created user.
func (r *Repo) Create(ctx context.Context, appleSub string) (*User, error) {
	const q = `
		INSERT INTO users (apple_sub)
		VALUES ($1)
		RETURNING id, apple_sub, COALESCE(username, ''), COALESCE(handle, ''), created_at`

	var u User
	err := r.db.QueryRow(ctx, q, appleSub).
		Scan(&u.ID, &u.AppleSub, &u.Username, &u.Handle, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpsertProfile updates username and handle for an existing user.
func (r *Repo) UpsertProfile(ctx context.Context, userID, username, handle string) (*User, error) {
	const q = `
		UPDATE users SET username = $1, handle = $2
		WHERE id = $3
		RETURNING id, apple_sub, COALESCE(username, ''), COALESCE(handle, ''), created_at`

	var u User
	err := r.db.QueryRow(ctx, q, username, handle, userID).
		Scan(&u.ID, &u.AppleSub, &u.Username, &u.Handle, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByID retrieves a user by primary key.
// Returns nil, nil when no user is found.
func (r *Repo) GetByID(ctx context.Context, userID string) (*User, error) {
	const q = `
		SELECT id, apple_sub, COALESCE(username, ''), COALESCE(handle, ''), created_at
		FROM users WHERE id = $1`

	var u User
	err := r.db.QueryRow(ctx, q, userID).
		Scan(&u.ID, &u.AppleSub, &u.Username, &u.Handle, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SaveDeviceToken persists an APNs device token for the given user.
func (r *Repo) SaveDeviceToken(ctx context.Context, userID, deviceToken string) error {
	const q = `UPDATE users SET device_token = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, q, userID, deviceToken)
	return err
}

// GetDeviceToken retrieves the APNs device token for the given user.
// Returns "", nil when no token is stored or the user is not found.
func (r *Repo) GetDeviceToken(ctx context.Context, userID string) (string, error) {
	const q = `SELECT COALESCE(device_token, '') FROM users WHERE id = $1`
	var token string
	err := r.db.QueryRow(ctx, q, userID).Scan(&token)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return token, nil
}

// SearchByHandle returns up to 20 users whose handle contains the given string (case-insensitive).
func (r *Repo) SearchByHandle(ctx context.Context, handle string) ([]User, error) {
	const q = `
		SELECT id, apple_sub, COALESCE(username, ''), COALESCE(handle, ''), created_at
		FROM users WHERE handle ILIKE $1 LIMIT 20`
	rows, err := r.db.Query(ctx, q, "%"+handle+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.AppleSub, &u.Username, &u.Handle, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
