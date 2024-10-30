package repo

import (
	"context"
	"database/sql"
	goerrors "errors"
	"strings"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/user"
)

const (
	getUserByID       = `SELECT id, username, password_hash, locality_id, created_at FROM MyUser WHERE id = $1`
	getUserByUsername = `SELECT id, username, password_hash, locality_id, created_at FROM MyUser WHERE username = $1`
	createUser        = `INSERT INTO MyUser (id, username, password_hash, locality_id, created_at) VALUES ($1, $2, $3, $4, $5)`
	setLocalityID     = `UPDATE MyUser SET locality_id = $1 WHERE id = $2`
)

type UserPostgres struct {
	db pgxtype.Querier
}

func NewUserPostgres(db pgxtype.Querier) *UserPostgres {
	return &UserPostgres{db: db}
}

func (repo *UserPostgres) GetUserByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	result := user.User{}
	var localityID []byte

	if err := repo.db.QueryRow(ctx, getUserByID, id).Scan(
		&result.ID,
		&result.Username,
		&result.PasswordHash,
		&localityID,
		&result.CreatedAt,
	); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return result, user.ErrUserNotFound
		}
		return result, errors.Wrap(err, "failed to get user from postgres")
	}

	result.LocalityID = uuid.FromBytesOrNil(localityID)
	return result, nil
}

func (repo *UserPostgres) GetUserByUsername(ctx context.Context, username string) (user.User, error) {
	result := user.User{}
	var localityID []byte

	if err := repo.db.QueryRow(ctx, getUserByUsername, username).Scan(
		&result.ID,
		&result.Username,
		&result.PasswordHash,
		&localityID,
		&result.CreatedAt,
	); err != nil {
		if goerrors.Is(err, pgx.ErrNoRows) {
			return result, user.ErrUserNotFound
		}
		return result, errors.Wrap(err, "failed to get user from postgres")
	}

	result.LocalityID = uuid.FromBytesOrNil(localityID)
	return result, nil
}

func (repo *UserPostgres) CreateUser(ctx context.Context, userData user.User) error {
	var localityID any = userData.LocalityID
	if localityID == uuid.Nil {
		localityID = sql.NullByte{}
	}

	if _, err := repo.db.Exec(ctx, createUser,
		userData.ID,
		userData.Username,
		userData.PasswordHash,
		localityID,
		userData.CreatedAt,
	); err != nil {
		if strings.HasSuffix(err.Error(), "(SQLSTATE 23505)") {
			return user.ErrUserAlreadyExists
		}
		return errors.Wrap(err, "failed to create user in postgres")
	}

	return nil
}

func (repo *UserPostgres) SetLocalityID(ctx context.Context, id uuid.UUID, localityID uuid.UUID) error {
	if _, err := repo.db.Exec(ctx, setLocalityID, id, localityID); err != nil {
		return errors.Wrap(err, "failed to set locality")
	}

	return nil
}
