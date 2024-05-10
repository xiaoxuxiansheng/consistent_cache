package example

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/xiaoxuxiansheng/consistent_cache"
	"github.com/xiaoxuxiansheng/consistent_cache/mysql"
	"github.com/xiaoxuxiansheng/consistent_cache/redis"
)

const (
	redisAddress  = "请输入 redis 地址"
	redisPassword = "请输入 redis 密码"

	mysqlDSN = "请输入 mysql dsn"
)

func newService() *consistent_cache.Service {
	// 缓存模块
	cache := redis.NewRedisCache(&redis.Config{
		Address:  redisAddress,
		Password: redisPassword,
	})
	// 数据库模块
	db := mysql.NewDB(mysqlDSN)
	return consistent_cache.NewService(cache, db,
		consistent_cache.WithCacheExpireSeconds(120),
		consistent_cache.WithDisableExpireSeconds(1),
	)
}

func Test_consistent_Cache(t *testing.T) {
	service := consistent_cache.NewService(
		// 缓存模块
		redis.NewRedisCache(&redis.Config{
			// redis 地址
			Address: redisAddress,
			// redis 密码
			Password: redisPassword,
		}),
		// 数据库模块
		mysql.NewDB(mysqlDSN),
		// 缓存过期时长 120s
		consistent_cache.WithCacheExpireSeconds(120),
		// 缓存过期时间添加随机扰动系数，防缓存雪崩
		consistent_cache.WithCacheExpireRandomMode(),
		// 写缓存禁用机制延时 1s 启用
		consistent_cache.WithDisableExpireSeconds(1),
	)
	ctx := context.Background()
	exp := Example{
		Key_: "test",
		Data: "test",
	}
	// 写操作
	if err := service.Put(ctx, &exp); err != nil {
		t.Error(err)
		return
	}

	// 读操作
	expReceiver := Example{
		Key_: "test",
	}
	if _, err := service.Get(ctx, &expReceiver); err != nil {
		t.Error(err)
		return
	}

	// 读取到的数据结果 以及是否使用到缓存
	t.Logf("read data: %s, ", expReceiver.Data)
}

// 验证点：1 数据正确性 2 缓存使用率
func Test_Consistent_Cache_Correct(t *testing.T) {
	// 构造缓存一致性服务实例
	service := newService()
	// 上下文、随机数生成器
	ctx := context.Background()
	rander := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 构造 500 个协程并发写，在本地备份一份写入的数据
	// 数据统一前缀
	prefix := time.Now().String() + "-"
	// 该 channel 用于接收来自写协程提交的数据，完成本地的冗余备份
	datac := make(chan *Example)
	// 异步启动 500 个协程并发写
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				k := prefix + cast.ToString(rander.Intn(100))
				v := prefix + cast.ToString(rander.Intn(100))
				data := Example{
					Key_: k,
					Data: v,
				}
				// 调用一致性服务进行写操作
				if err := service.Put(ctx, &data); err != nil {
					t.Error(err)
					return
				}
				// 写成功后，数据通过 channel 发往本地读协程，进行数据备份
				datac <- &data
			}()
		}
		wg.Wait()
		close(datac)
	}()

	// 通过 channel 接收来自写协程提交的数据，在本地完成数据备份
	mp := make(map[string]string, 500)
	for data := range datac {
		mp[data.Key_] = data.Data
	}

	// 缓冲一秒，等待写操作的 disable 操作过期
	<-time.After(time.Second)

	// 记录读操作命中缓存的次数
	var useCacheCnt int
	// 预期读操作命中缓存的次数
	var expectUseCacheCnt int
	querySet := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		k := cast.ToString(rander.Intn(100))
		data := Example{
			Key_: prefix + k,
		}
		if _, ok := querySet[prefix+k]; ok {
			expectUseCacheCnt++
		}
		querySet[prefix+k] = struct{}{}

		// 通过一致性缓存服务发起读操作
		useCache, err := service.Get(ctx, &data)
		if err != nil && !errors.Is(err, consistent_cache.ErrorDataNotExist) {
			t.Error(err)
			continue
		}

		if useCache {
			useCacheCnt++
		}

		// 利用本地备份数据和读取结果进行对比校验
		expect, ok := mp[data.Key_]
		assert.Equal(t, !ok, errors.Is(err, consistent_cache.ErrorDataNotExist))
		if !ok {
			continue
		}

		assert.Equal(t, expect, data.Data)
	}

	// 校验操作命中读缓存是否与预期一致
	assert.Equal(t, expectUseCacheCnt, useCacheCnt)
}

// 读写操作并发执行 验证点 1：disable 机制正常启用 2：读取结果正确
func Test_Consistent_Cache_Read_Write(t *testing.T) {
	// 构造缓存一致性服务实例
	service := newService()

	ctx := context.Background()

	// 数据统一前缀
	prefix := time.Now().String()

	// 并发控制、协程间数据传递
	var wg sync.WaitGroup
	datac := make(chan *Example)

	// value 值范围
	startV, endV := 1, 5
	// 启动多个协程写同一个 key，value 值取在 [startV,endV] 之间
	go func() {
		for i := startV; i <= endV; i++ {
			i := i // shadow
			wg.Add(1)
			go func() {
				defer wg.Done()
				k := prefix
				v := prefix + cast.ToString(i)
				data := Example{
					Key_: k,
					Data: v,
				}
				// 调用一致性服务进行写操作
				if err := service.Put(ctx, &data); err != nil {
					t.Error(err)
				}
				// 写成功后，数据通过 channel 发往本地读协程，进行数据备份
				datac <- &data
			}()
		}
	}()

	// 启动双倍的读协程数量，读同一个 key
	go func() {
		for i := 0; i < 10*(endV-startV+1); i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				data := Example{
					Key_: prefix,
				}
				// 调用一致性服务进行写操作
				useCache, err := service.Get(ctx, &data)
				if err != nil && !errors.Is(err, consistent_cache.ErrorDataNotExist) {
					t.Error(err)
					return
				}
				if errors.Is(err, consistent_cache.ErrorDataNotExist) {
					return
				}
				// 预期不能用上缓存
				assert.Equal(t, false, useCache)
				// 数据预期是 0-4 之一都有可能
				gotData := cast.ToInt(data.Data)
				assert.Equal(t, true, gotData >= startV && gotData <= endV)
			}()
		}
	}()

	// 通过 channel 接收来自写协程提交的数据，在本地完成数据备份
	datas := make([]*Example, 0, 5)
	for i := startV; i <= endV; i++ {
		data := <-datac
		datas = append(datas, data)
	}

	// 尘埃落定后，读取到最终的正确结果
	// 调用一致性服务进行写操作
	data := Example{
		Key_: prefix,
	}
	useCache, err := service.Get(ctx, &data)
	if err != nil {
		t.Error(err)
		return
	}
	// 预期不能用上缓存
	assert.Equal(t, false, useCache)
	// 预期结果为最后一笔写入的数据
	assert.Equal(t, datas[len(datas)-1].Data, data.Data)

	wg.Wait()

	// 1秒后，重复读取两次，第一次不命中缓存，第二次命中缓存
	<-time.After(time.Second)
	if useCache, err = service.Get(ctx, &data); err != nil {
		t.Error(err)
		return
	}
	// 预期不能用上缓存
	assert.Equal(t, false, useCache)

	if useCache, err = service.Get(ctx, &data); err != nil {
		t.Error(err)
		return
	}
	// 第二次行为预期用上缓存
	assert.Equal(t, true, useCache)
	// 读取结果应该等同于最晚一笔写入的内容
	assert.Equal(t, datas[len(datas)-1].Data, data.Data)
}
