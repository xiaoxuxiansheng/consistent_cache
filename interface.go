package consistent_cache

import (
	"context"
	"errors"
)

var (
	ErrorDataNotExist = errors.New("data not exist")
	ErrorCacheMiss    = errors.New("cache miss")
	ErrorDBMiss       = errors.New("db miss")
)

const NullData = "Err_Syntax_Null_Data"

// 缓存模块的抽象接口定义
type Cache interface {
	// 启用某个 key 对应读流程写缓存机制（默认情况下为启用状态）
	Enable(ctx context.Context, key string, delayMilis int64) error
	// 禁用某个 key 对应读流程写缓存机制
	Disable(ctx context.Context, key string, expireSeconds int64) error
	// 读取 key 对应缓存
	Get(ctx context.Context, key string) (string, error)
	// 删除 key 对应缓存
	Del(ctx context.Context, key string) error
	// 校验某个 key 对应读流程写缓存机制是否启用，倘若启用则写入缓存（默认情况下为启用状态）
	PutWhenEnable(ctx context.Context, key, value string, expireSeconds int64) (bool, error)
}

// 数据库模块的抽象接口定义
type DB interface {
	// 数据写入数据库
	Put(ctx context.Context, obj Object) error
	// 从数据库读取数据
	Get(ctx context.Context, obj Object) error
}

type Object interface {
	// 获取 key 对应的字段名
	KeyColumn() string
	// 获取 key 对应的值
	Key() string
	// 数据对应的字段名
	DataColumn() []string

	// 将 object 序列化成字符串
	Write() (string, error)
	// 读取字符串内容，反序列化到 object 实例中
	Read(body string) error
}

type Logger interface {
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}
