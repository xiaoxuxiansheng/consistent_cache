package consistent_cache

import (
	"context"
	"errors"
	"time"
)

// 一致性缓存服务
type Service struct {
	// 配置项
	opts *Options
	// 缓存模块
	cache Cache
	// 数据库模块
	db DB
}

// 构造一致性缓存服务. 缓存和数据库均由使用方提供具体的实现版本
func NewService(cache Cache, db DB, opts ...Option) *Service {
	s := Service{
		cache: cache,
		db:    db,
		opts:  &Options{},
	}

	for _, opt := range opts {
		opt(s.opts)
	}

	repair(s.opts)
	return &s
}

// 写操作
func (s *Service) Put(ctx context.Context, obj Object) error {
	// 1 针对 key 维度禁用读流程写缓存机制
	if err := s.cache.Disable(ctx, obj.Key(), s.opts.disableExpireSeconds); err != nil {
		return err
	}

	defer func() {
		go func() {
			tctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := s.cache.Enable(tctx, obj.Key(), s.opts.enableDelayMilis); err != nil {
				s.opts.logger.Errorf("enable fail, key: %s, err: %v", obj.Key(), err)
			}
		}()
	}()

	// 2 删除 key 维度对应缓存
	if err := s.cache.Del(ctx, obj.Key()); err != nil {
		return err
	}

	// 3 数据写入 db
	return s.db.Put(ctx, obj)
}

// 2 读操作
func (s *Service) Get(ctx context.Context, obj Object) (useCache bool, err error) {
	// 1 读取缓存
	v, err := s.cache.Get(ctx, obj.Key())
	// 2 非缓存 miss 类错误，直接抛出错误
	if err != nil && !errors.Is(err, ErrorCacheMiss) {
		return false, err
	}

	// 3 读取到缓存结果
	if err == nil {
		// 3.1 读取到的数据为 EmptyData. 是为了防止缓存穿透而设置的空值
		if v == NullData {
			return true, ErrorDataNotExist
		}
		// 3.2 正常读取到数据
		return true, obj.Read(v)
	}

	// 4 缓存 miss，读 db
	if err = s.db.Get(ctx, obj); err != nil && !errors.Is(err, ErrorDBMiss) {
		return false, err
	}

	// 5 db 中也没有数据，则尝试往 cache 中写入 NullData
	if errors.Is(err, ErrorDBMiss) {
		if ok, err := s.cache.PutWhenEnable(ctx, obj.Key(), NullData, s.opts.CacheExpireSeconds()); err != nil {
			s.opts.logger.Errorf("put null data into cache fail, key: %s, err: %v", obj.Key(), err)
		} else {
			s.opts.logger.Infof("put null data into cache resp, key: %s, ok: %t", obj.Key(), ok)
		}

		return false, ErrorDataNotExist
	}

	// 6 成功获取到数据了，则需要将其写入缓存
	v, err = obj.Write()
	if err != nil {
		return false, err
	}
	if ok, err := s.cache.PutWhenEnable(ctx, obj.Key(), v, s.opts.CacheExpireSeconds()); err != nil {
		s.opts.logger.Errorf("put data into cache fail, key: %s, data: %v, err: %v", obj.Key(), v, err)
	} else {
		s.opts.logger.Infof("put data into cache resp, key: %s, v: %v, ok: %t", obj.Key(), v, ok)
	}

	// 7 返回读取到的结果
	return false, nil
}
