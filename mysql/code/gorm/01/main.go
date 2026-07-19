package main

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 连接数据库

func main() {
	// DSN（Data Source Name）格式：
	// 用户名:密码@tcp(主机:端口)/数据库名?参数
	//
	// parseTime=true：将 MySQL 的 DATE、DATETIME 等类型解析为 time.Time。
	// loc=Local：使用当前系统所在时区解析时间。
	dsn := "root:123@tcp(127.0.0.1:3306)/learn_mysql?charset=utf8mb4&parseTime=true&loc=Local"

	// 创建 GORM 数据库对象。
	// 默认情况下，会自动执行 Ping 来检查数据库连接是否可用
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 MySQL 失败：%v", err)
	}
	log.Println("MySQL 连接成功")

	// 获取底层的连接池。
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取连接池失败：%v", err)
	}
	defer sqlDB.Close()

	// 配置连接池。

	// 最大打开连接数，包括正在使用和空闲的连接。
	sqlDB.SetMaxOpenConns(20)

	// 最大空闲连接数。
	sqlDB.SetMaxIdleConns(10)

	// 每个连接最多可以复用一小时，之后会被关闭并重新创建。
	sqlDB.SetConnMaxLifetime(time.Hour)
}
