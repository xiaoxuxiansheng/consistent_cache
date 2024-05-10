package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/xiaoxuxiansheng/consistent_cache"
)

type tabler interface {
	TableName() string
}

// 数据库模块的抽象接口定义
type DB struct {
	db *gorm.DB
}

func NewDB(dsn string) *DB {
	return &DB{db: getDB(dsn)}
}

// 数据写入数据库
func (d *DB) Put(ctx context.Context, obj consistent_cache.Object) error {
	db := d.db
	tabler, ok := obj.(tabler)
	if ok {
		db = db.Table(tabler.TableName())
	}

	// upsert 操作
	err := db.WithContext(ctx).Create(obj).Error
	if err == nil {
		return nil
	}
	// 是否为唯一键冲突
	flag := IsDuplicateEntryErr(err)
	// 唯一键冲突，则改为更新操作
	if flag {
		return db.WithContext(ctx).Debug().Where(fmt.Sprintf("`%s` = ?", obj.KeyColumn()), obj.Key()).Updates(obj).Error
	}
	// 其他错误直接返回
	return err
}

// 从数据库读取数据
func (d *DB) Get(ctx context.Context, obj consistent_cache.Object) error {
	db := d.db
	tabler, ok := obj.(tabler)
	if ok {
		db = db.Table(tabler.TableName())
	}

	err := db.WithContext(ctx).Where(fmt.Sprintf("`%s` = ?", obj.KeyColumn()), obj.Key()).First(obj).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return consistent_cache.ErrorDBMiss
	}
	return err
}
