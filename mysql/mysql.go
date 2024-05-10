package mysql

import (
	"errors"

	mysql2 "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 唯一键冲突错误
const DuplicateEntryErrCode = 1062

// GetClient 获取一个数据库客户端
func getDB(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

// 是否为唯一键冲突错误
func IsDuplicateEntryErr(err error) bool {
	var mysqlErr *mysql2.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == DuplicateEntryErrCode {
		return true
	}
	return false
}
