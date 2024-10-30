package repo

import (
	"context"
	goerrors "errors"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type SessionRedis struct {
	client *redis.Client
}

func NewSessionRedis(client *redis.Client) *SessionRedis {
	return &SessionRedis{client: client}
}

func (s *SessionRedis) SetAccessToken(ctx context.Context, username string, token string, lifeTime time.Duration) error {
	if err := s.client.Set(ctx, getAccessTokenKey(username), token, lifeTime).Err(); err != nil {
		return errors.Wrap(err, "failed to set access token")
	}

	return nil
}

func (s *SessionRedis) SetRefreshToken(ctx context.Context, username string, token string) error {
	if err := s.client.Set(ctx, getRefreshTokenKey(username), token, 0).Err(); err != nil {
		return errors.Wrap(err, "failed to set refresh token")
	}

	return nil
}

func (s *SessionRedis) GetAccessToken(ctx context.Context, username string) (string, error) {
	token, err := s.client.Get(ctx, getAccessTokenKey(username)).Result()
	if err != nil {
		if goerrors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", errors.Wrap(err, "failed to get access token")
	}

	return token, nil
}

func (s *SessionRedis) GetRefreshToken(ctx context.Context, username string) (string, error) {
	token, err := s.client.Get(ctx, getRefreshTokenKey(username)).Result()
	if err != nil {
		if goerrors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", errors.Wrap(err, "failed to get refresh token")
	}

	return token, nil
}

func (s *SessionRedis) RemoveAccessToken(ctx context.Context, username string) error {
	if err := s.client.Del(ctx, getAccessTokenKey(username)).Err(); err != nil {
		return errors.Wrap(err, "failed to remove access token")
	}

	return nil
}

func (s *SessionRedis) RemoveRefreshToken(ctx context.Context, username string) error {
	if err := s.client.Del(ctx, getRefreshTokenKey(username)).Err(); err != nil {
		return errors.Wrap(err, "failed to remove refresh token")
	}

	return nil
}

func getAccessTokenKey(username string) string {
	return fmt.Sprintf("access:%s", username)
}

func getRefreshTokenKey(username string) string {
	return fmt.Sprintf("refresh:%s", username)
}
