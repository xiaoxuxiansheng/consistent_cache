package redis

import (
	"context"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Config struct {
	Address            string
	Password           string
	MaxIdle            int
	IdleTimeoutSeconds int
	// 连接池最大存活的连接数.
	MaxActive int
	// 当连接数达到上限时，新的请求是等待还是立即报错.
	Wait bool
}

type RClient struct {
	pool *redis.Pool
}

func NewRClient(config *Config) *RClient {
	return &RClient{
		pool: getRedisPool(config),
	}
}

func getRedisPool(config *Config) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     config.MaxIdle,
		IdleTimeout: time.Duration(config.IdleTimeoutSeconds) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := newRedisConn(config)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		MaxActive: config.MaxActive,
		Wait:      config.Wait,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func newRedisConn(conf *Config) (redis.Conn, error) {
	if conf.Address == "" {
		panic("Cannot get redis address from config")
	}

	conn, err := redis.Dial("tcp", conf.Address, redis.DialPassword(conf.Password))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *RClient) Get(ctx context.Context, key string) (string, error) {
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

func (r *RClient) SetEx(ctx context.Context, key, value string, expireSeconds int64) error {
	if key == "" {
		return errors.New("redis SET EX key can't be empty")
	}
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("SET", key, value, "EX", expireSeconds)
	return err
}

func (r *RClient) Del(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("redis DEL key can't be empty")
	}
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("DEL", key)
	return err
}

// Eval 支持使用 lua 脚本.
func (r *RClient) Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error) {
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

func (r *RClient) PExpire(ctx context.Context, key string, expireMilis int64) error {
	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Do("PEXPIRE", key, expireMilis)
	return err
}
