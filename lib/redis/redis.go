package redis

import (
	"context"
	"errors"

	"github.com/gomodule/redigo/redis"
)

type RedisClient struct {
	pool *redis.Pool
}

func NewRedisClient(pool *redis.Pool) *RedisClient {
	return &RedisClient{
		pool: pool,
	}
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", errors.New("redis GET key can't be empty")
	}
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return redis.String(conn.Do("GET", key))
}

// Eval 支持使用 lua 脚本.
func (r *RedisClient) Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error) {
	args := make([]interface{}, 2+len(keysAndArgs))
	args[0] = src
	args[1] = keyCount
	copy(args[2:], keysAndArgs)

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.Do("EVAL", args...)
}
