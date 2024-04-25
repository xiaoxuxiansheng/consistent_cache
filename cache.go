package consistent_cache

import "context"

type CacheClient interface {
	Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error)
}

type Cache struct {
	client CacheClient
}

func NewCache(client CacheClient) *Cache {
	return &Cache{client: client}
}

// 启用某个的 key 的读流程写缓存机制
func (c *Cache) Enable(key string) error {
	return nil
}

// 禁用某个的 key 的读流程写缓存机制
func (c *Cache) Disable(key string) error {
	return nil
}

// 读缓存
func (c *Cache) Get(key string) (string, error) {
	return "", nil
}

func (c *Cache) PutWhenEnable(key, value string) error {
	return nil
}

func (c *Cache) Put(key, value string) error {
	return nil
}

func (c *Cache) Del(key string) error {
	return nil
}
