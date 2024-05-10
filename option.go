package consistent_cache

import (
	"math/rand"
	"time"

	"github.com/xiaoxuxiansheng/consistent_cache/lib/log"
)

type Options struct {
	// 缓存过期时间，单位：秒
	cacheExpireSeconds int64
	// 是否启用过期时间扰动
	cacheExpireRandomMode bool
	// 禁用读流程写缓存模式过期时间，单位：秒
	disableExpireSeconds int64
	// 写流程 disable 操作后延时多长时间进行 enable 操作，单位：毫秒
	enableDelayMilis int64
	// 随机数生成器
	rander *rand.Rand
	// 日志打印
	logger Logger
}

func (o *Options) CacheExpireSeconds() int64 {
	if !o.cacheExpireRandomMode {
		return o.cacheExpireSeconds
	}

	// 过期时间在 1~2倍之间取随机值
	return o.cacheExpireSeconds + o.rander.Int63n(o.cacheExpireSeconds+1)
}

type Option func(*Options)

const (
	// 默认的缓存过期时间为 60 s
	DefaultCacheExpireSeconds = 60
	// 默认的禁用写缓存时间为 10 s
	DefaultDisableExpireSeconds = 10
	// 默认的延时 enable 时间为 1 s
	DefaultEnableDelayMilis = 1000
)

func WithCacheExpireSeconds(cacheExpireSeconds int64) Option {
	return func(o *Options) {
		o.cacheExpireSeconds = cacheExpireSeconds
	}
}

func WithCacheExpireRandomMode() Option {
	return func(o *Options) {
		o.cacheExpireRandomMode = true
		o.rander = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
}

func WithDisableExpireSeconds(disableExpireSeconds int64) Option {
	return func(o *Options) {
		o.disableExpireSeconds = disableExpireSeconds
	}
}

func WithEnableDelayMilis(enableDelayMilis int64) Option {
	return func(o *Options) {
		o.enableDelayMilis = enableDelayMilis
	}
}

func WithLogger(logger Logger) Option {
	return func(o *Options) {
		o.logger = logger
	}
}

func repair(o *Options) {
	if o.cacheExpireSeconds <= 0 {
		o.cacheExpireSeconds = DefaultCacheExpireSeconds
	}

	if o.disableExpireSeconds <= 0 {
		o.disableExpireSeconds = DefaultDisableExpireSeconds
	}

	if o.enableDelayMilis <= 0 {
		o.enableDelayMilis = DefaultEnableDelayMilis
	}

	if o.logger == nil {
		o.logger = log.GetLogger()
	}
}
