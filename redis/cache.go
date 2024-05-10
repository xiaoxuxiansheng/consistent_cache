package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/spf13/cast"
	"github.com/xiaoxuxiansheng/consistent_cache"
)

// redis 客户端.
type Client interface {
	Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error)
	Get(ctx context.Context, key string) (string, error)
	SetEx(ctx context.Context, key, value string, expireSeconds int64) error
	Del(ctx context.Context, key string) error
	PExpire(ctx context.Context, key string, expireMilis int64) error
}

// redis 实现版本的缓存模块
type Cache struct {
	client Client
}

// 构造器函数
func NewRedisCache(config *Config) *Cache {
	return &Cache{client: NewRClient(config)}
}

// 启用某个 key 对应读流程写缓存机制（默认情况下为启用状态）
func (c *Cache) Enable(ctx context.Context, key string, delayMilis int64) error {
	// redis 中删除 key 对应的 disable key. 只要 disable key 标识不存在，则读流程写缓存机制视为启用状态
	// 给 disable key 设置一个相对较短的过期时间
	return c.client.PExpire(ctx, key, delayMilis)
}

// 禁用某个 key 的读流程写缓存机制
func (c *Cache) Disable(ctx context.Context, key string, expireSeconds int64) error {
	// redis 中设置 key 对应的 disable key. 只要 disable key 标识存在，则读流程写缓存机制视为禁用状态
	return c.client.SetEx(ctx, c.disableKey(key), "1", expireSeconds)
}

// 读取 key 对应缓存内容
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	// 从 redis 中读取 kv 对
	reply, err := c.client.Get(ctx, key)
	if err != nil && !errors.Is(err, redis.ErrNil) {
		return "", err
	}
	if errors.Is(err, redis.ErrNil) {
		return "", consistent_cache.ErrorCacheMiss
	}
	return reply, nil
}

// 校验某个 key 对应读流程写缓存机制是否启用，倘若启用则写入缓存（默认情况下为启用状态）
func (c *Cache) PutWhenEnable(ctx context.Context, key, value string, expireSeconds int64) (bool, error) {
	// 运行 redis lua 脚本，保证只有在 disable key 不存在时，才会执行 key 的写入
	reply, err := c.client.Eval(ctx, LuaCheckEnableAndWriteCache, 2, []interface{}{
		c.disableKey(key),
		key,
		value,
		expireSeconds,
	})
	if err != nil {
		return false, err
	}
	return cast.ToInt(reply) == 1, nil
}

// 删除 key 对应缓存
func (c *Cache) Del(ctx context.Context, key string) error {
	// 从 reids 中删除 kv 对
	return c.client.Del(ctx, key)
}

// 基于 key 映射得到 v key 表达式
func (c *Cache) disableKey(key string) string {
	// 通过 {hash_tag}，保证在 redis 集群模式下，key 和 disable key 也会被分发到相同节点
	return fmt.Sprintf("Enable_Lock_Key_{%s}", key)
}
