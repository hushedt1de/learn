package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	// go get github.com/go-sql-driver/mysql 安装 MySQL 驱动。
	// 通过匿名导入注册 MySQL 驱动。
	_ "github.com/go-sql-driver/mysql"
)

// 连接数据库

func main() {
	// DSN（Data Source Name）格式：
	// 用户名:密码@tcp(主机:端口)/数据库名?参数
	//
	// parseTime=true：将 MySQL 的 DATE、DATETIME 等类型解析为 time.Time。
	// loc=Local：使用当前系统所在时区解析时间。
	dsn := "root:123@tcp(127.0.0.1:3306)/learn_mysql?charset=utf8mb4&parseTime=true&loc=Local"

	// 创建一个并发安全的数据库连接池。通常在程序启动时创建，并在整个程序运行期间复用。
	// 注意：这里通常不会立即连接 MySQL，因此不能仅凭 err 判断数据库是否可用。
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("创建数据库连接池失败：%v", err)
	}
	// 程序退出时关闭连接池。
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// PingContext 会真正尝试连接 MySQL，用来确认数据库是否可用。
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("连接 MySQL 失败：%v", err)
	}
	log.Println("MySQL 连接成功")

	// 配置连接池。

	// 最大打开连接数，包括正在使用和空闲的连接。
	db.SetMaxOpenConns(20)

	// 最大空闲连接数。
	db.SetMaxIdleConns(10)

	// 每个连接最多可以复用一小时，之后会被关闭并重新创建。
	db.SetConnMaxLifetime(time.Hour)
}
