package auth

import "context"

type UserRepo interface {
	CreateUser(ctx context.Context, username, password string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByID(ctx context.Context, userID UserID) (*User, error)
}
