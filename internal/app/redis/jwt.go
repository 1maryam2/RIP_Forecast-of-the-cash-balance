package redis

import (
	"context"
	"strconv"
	"time"
)

const jwtPrefix = "jwt."

func getJWTKey(token string) string {
	return servicePrefix + jwtPrefix + token
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, jwtTTL time.Duration) error {
	return c.client.Set(ctx, getJWTKey(jwtStr), true, jwtTTL).Err()
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) error {
	return c.client.Get(ctx, getJWTKey(jwtStr)).Err()
}

const refreshTokenPrefix = "refresh_token."

func getRefreshTokenKey(userUUID string) string {
	return servicePrefix + refreshTokenPrefix + userUUID
}

func (c *Client) StoreRefreshToken(ctx context.Context, userUUID, token string, ttl time.Duration) error {
	return c.client.Set(ctx, getRefreshTokenKey(userUUID), token, ttl).Err()
}

func (c *Client) GetRefreshToken(ctx context.Context, userUUID string) (string, error) {
	return c.client.Get(ctx, getRefreshTokenKey(userUUID)).Result()
}

func (c *Client) DeleteRefreshToken(ctx context.Context, userUUID string) error {
	return c.client.Del(ctx, getRefreshTokenKey(userUUID)).Err()
}

const userSessionPrefix = "user_session."

func getUserSessionKey(userID uint) string {
	return servicePrefix + userSessionPrefix + strconv.FormatUint(uint64(userID), 10)
}

func (c *Client) StoreUserSession(ctx context.Context, userID uint, data string, ttl time.Duration) error {
	return c.client.Set(ctx, getUserSessionKey(userID), data, ttl).Err()
}

func (c *Client) GetUserSession(ctx context.Context, userID uint) (string, error) {
	return c.client.Get(ctx, getUserSessionKey(userID)).Result()
}

func (c *Client) DeleteUserSession(ctx context.Context, userID uint) error {
	return c.client.Del(ctx, getUserSessionKey(userID)).Err()
}
