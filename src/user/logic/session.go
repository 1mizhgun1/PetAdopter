package logic

import (
	"context"

	"github.com/pkg/errors"
	"pet_adopter/src/config"
	"pet_adopter/src/user"
	"pet_adopter/src/utils"
)

type SessionLogic struct {
	session user.SessionRepo
	cfg     config.SessionConfig
}

func NewSessionLogic(session user.SessionRepo, cfg config.SessionConfig) *SessionLogic {
	return &SessionLogic{
		session: session,
		cfg:     cfg,
	}
}

func (logic *SessionLogic) CheckSession(ctx context.Context, username string, token string) (bool, error) {
	accessToken, err := logic.session.GetAccessToken(ctx, username)
	if err != nil {
		return false, errors.Wrap(err, "failed to get access token")
	}

	return token == accessToken, nil
}

func (logic *SessionLogic) SetSession(ctx context.Context, username string) (string, string, error) {
	accessToken := utils.GenerateSessionToken(logic.cfg.AccessTokenLength)
	if err := logic.session.SetAccessToken(ctx, username, accessToken, logic.cfg.AccessTokenLifeTime); err != nil {
		return "", "", errors.Wrap(err, "failed to set access token")
	}

	refreshToken := utils.GenerateSessionToken(logic.cfg.RefreshTokenLength)
	if err := logic.session.SetRefreshToken(ctx, username, refreshToken); err != nil {
		return "", "", errors.Wrap(err, "failed to set refresh token")
	}

	return accessToken, refreshToken, nil
}

func (logic *SessionLogic) RefreshSession(ctx context.Context, username string, refreshToken string) (string, string, error) {
	token, err := logic.session.GetRefreshToken(ctx, username)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get refresh token")
	}

	if refreshToken != token {
		return "", "", user.ErrInvalidRefreshToken
	}

	return logic.SetSession(ctx, username)
}
