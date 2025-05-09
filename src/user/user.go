package user

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrUserAlreadyExists   = errors.New("user already exists")
)

type User struct {
	ID           uuid.UUID `json:"-"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	LocalityID   uuid.UUID `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

type UserRepo interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	CreateUser(ctx context.Context, user User) error
	SetLocalityID(ctx context.Context, id uuid.UUID, localityID uuid.UUID) error
}

type SessionRepo interface {
	SetAccessToken(ctx context.Context, username string, token string, lifeTime time.Duration) error
	SetRefreshToken(ctx context.Context, username string, token string) error
	GetAccessToken(ctx context.Context, username string) (string, error)
	GetRefreshToken(ctx context.Context, username string) (string, error)
	RemoveAccessToken(ctx context.Context, username string) error
	RemoveRefreshToken(ctx context.Context, username string) error
}

type UserLogic interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
	CreateUser(ctx context.Context, username string, password string) (User, error)
	SetLocalityID(ctx context.Context, id uuid.UUID, localityID uuid.UUID) (User, error)
	CheckPassword(ctx context.Context, username string, password string) (User, bool, error)
}

type SessionLogic interface {
	CheckSession(ctx context.Context, username string, token string) (bool, error)
	SetSession(ctx context.Context, username string) (string, string, error)
}
