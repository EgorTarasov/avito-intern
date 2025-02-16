package storage

import (
	"avito-intern/internal/auth"
	"avito-intern/pkg/db"
	"context"
	"fmt"
	"time"

	"errors"

	"github.com/jackc/pgx/v5"
)

type pgUser struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
}

// PgUserRepo is a repository for PostgreSQL.
type PgUserRepo struct {
	database *db.Database
}

// NewPgUserRepo creates a new PgUserRepo instance.
func NewPgUserRepo(database *db.Database) *PgUserRepo {
	return &PgUserRepo{
		database: database,
	}
}

func mapUser(user *pgUser) *auth.User {
	return &auth.User{
		ID:        auth.UserID(user.ID),
		Username:  user.Username,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
	}
}

// CreateUser creates a new user and returns it.
// It assumes a table "users" with columns id, username, and password.
func (r *PgUserRepo) CreateUser(ctx context.Context, username, password string) (*auth.User, error) {
	query := `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id, username, password`
	var user pgUser
	if err := r.database.Get(ctx, &user, query, username, password); err != nil {
		// if no row is returned, consider it as not found.
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return mapUser(&user), nil
}

// GetUserByUsername returns the user matching the specified username.
func (r *PgUserRepo) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	query := `SELECT id, username, password FROM users WHERE username = $1`
	var user pgUser
	if err := r.database.Get(ctx, &user, query, username); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return mapUser(&user), nil
}

// GetUserByID returns the user with the given ID.
func (r *PgUserRepo) GetUserByID(ctx context.Context, userID auth.UserID) (*auth.User, error) {
	query := `SELECT id, username, password FROM users WHERE id = $1`
	var user pgUser
	if err := r.database.Get(ctx, &user, query, int64(userID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return mapUser(&user), nil
}
