package logic

import (
	"context"
	goerrors "errors"
	"time"

	"github.com/pkg/errors"
	"github.com/satori/uuid"
	"pet_adopter/src/locality"
	"pet_adopter/src/user"
	"pet_adopter/src/utils"
)

type UserLogic struct {
	repo         user.UserRepo
	localityRepo locality.LocalityRepo
}

func NewUserLogic(repo user.UserRepo, localityRepo locality.LocalityRepo) *UserLogic {
	return &UserLogic{repo: repo, localityRepo: localityRepo}
}

func (logic *UserLogic) GetUserByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	return logic.repo.GetUserByID(ctx, id)
}

func (logic *UserLogic) GetUserByUsername(ctx context.Context, username string) (user.User, error) {
	return logic.repo.GetUserByUsername(ctx, username)
}

func (logic *UserLogic) CreateUser(ctx context.Context, username string, password string) (user.User, error) {
	userData := user.User{
		ID:           uuid.NewV4(),
		Username:     username,
		PasswordHash: utils.GetPasswordHash(password),
		LocalityID:   uuid.Nil,
		CreatedAt:    time.Now().Local(),
	}

	if err := logic.repo.CreateUser(ctx, userData); err != nil {
		return user.User{}, err
	}

	return userData, nil
}

func (logic *UserLogic) SetLocalityID(ctx context.Context, id uuid.UUID, localityID uuid.UUID) (user.User, error) {
	_, err := logic.localityRepo.GetLocalityByID(ctx, localityID)
	if err != nil {
		if goerrors.Is(err, locality.ErrLocalityNotFound) {
			return user.User{}, locality.ErrLocalityNotFound
		}
		return user.User{}, errors.Wrap(err, "failed to get locality")
	}

	if err = logic.repo.SetLocalityID(ctx, id, localityID); err != nil {
		return user.User{}, errors.Wrap(err, "failed to set locality id")
	}

	return logic.repo.GetUserByID(ctx, id)
}

func (logic *UserLogic) CheckPassword(ctx context.Context, username string, password string) (user.User, bool, error) {
	userData, err := logic.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return user.User{}, false, errors.Wrap(err, "failed to get user data")
	}

	return userData, utils.GetPasswordHash(password) == userData.PasswordHash, nil
}
