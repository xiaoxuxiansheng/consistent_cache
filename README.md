<p align="center">
<img src="https://github.com/xiaoxuxiansheng/consistent_cache/blob/main/img/frame.png" />
<b>goredis: 基于 go 实现缓存读写一致性服务</b>
<br/><br/>
</p>

## 📖 简介
本着学习和实践的目标, 基于 100% 纯度 go 语言实现的缓存读写一致性 lib 框架，实现到的功能点包括：
- 缓存一致性保证 
    - 写流程: 设置禁用写缓存标识 -> 删除缓存 -> 写数据库 -> 延时启用写缓存标识
<img src="https://github.com/xiaoxuxiansheng/consistent_cache/blob/main/img/write_process.png" />
    - 读流程: 读缓存 -> 读数据库 -> 仅在写缓存标识启用时写缓存
<img src="https://github.com/xiaoxuxiansheng/consistent_cache/blob/main/img/read_process.png" />
- 缓存雪崩防治
    - 针对缓存过期时间添加随机扰动 防止海量数据同时刻过期
- 缓存穿透对策
    - 缓存中添加 NullData 防止不存在数据发生缓存穿透问题

## 💡 技术原理分享
<a href="">一致性缓存理论分析与技术实战(待补充链接)</a> <br/><br/>

## 🖥 使用示例
```go
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
```