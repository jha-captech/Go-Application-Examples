package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient is an interface that defines some of the methods used by Redis.
type RedisClient interface {
	Set(
		ctx context.Context,
		key string,
		value interface{},
		expiration time.Duration,
	) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
}

// Wraps a *redis.Client and provides methods for setting and getting cached values with
// automatic JSON marshaling/unmarshaling and expiration handling.
type Client struct {
	Redis      RedisClient
	expiration time.Duration
}

// Constructs a new Client with the given Redis client and default expiration duration.
func NewClient(client RedisClient, expiration time.Duration) *Client {
	return &Client{
		Redis:      client,
		expiration: expiration,
	}
}

// Wraps redis.StringCmd to provide additional helpers for result extraction and unmarshaling.
type StringCmd struct {
	*redis.StringCmd
}

// Marshals the given value to JSON and stores it in Redis under the specified key with the
// configured expiration.
func (c *Client) SetMarshal(ctx context.Context, key string, value any) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("[in services.Client.Set] failed to marshal value: %w", err)
	}

	if err = c.Redis.Set(ctx, key, jsonData, c.expiration).Err(); err != nil {
		return fmt.Errorf("[in services.Client.Set] failed to set value in cache: %w", err)
	}

	return nil
}

// Retrieves the value for the given key from Redis, returning a StringCmd wrapper.
func (c *Client) Get(ctx context.Context, key string) *StringCmd {
	return &StringCmd{c.Redis.Get(ctx, key)}
}

// Deletes the value for the given key from Redis, returning an error if any.
func (c *Client) Delete(ctx context.Context, key string) error {
	return c.Redis.Del(ctx, key).Err()
}

// Returns the string result, a boolean indicating existence, and an error if any.
func (cmd *StringCmd) Result() (string, bool, error) {
	val, err := cmd.StringCmd.Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return "", false, nil
		default:
			return "", false, err
		}
	}

	if val == "" {
		return "", false, nil
	}

	return val, true, nil
}

// Unmarshals the JSON value from Redis into the provided variable. Returns a boolean indicating
// existence and an error if unmarshaling fails.
func (cmd *StringCmd) Unmarshal(v any) (bool, error) {
	val, err := cmd.StringCmd.Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return false, nil
		default:
			return false, err
		}
	}

	if val == "" {
		return false, nil
	}

	if err = json.Unmarshal([]byte(val), v); err != nil {
		return false, fmt.Errorf(
			"[in services.StringCmd.Unmarshal] failed to unmarshal from cache: %w",
			err,
		)
	}

	return true, nil
}
