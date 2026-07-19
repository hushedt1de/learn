package main

import (
	"context"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 创建和删除数据库
// 生产环境一般不建议让业务服务随意创建或删除数据库，更常见的是通过部署脚本或管理员账号执行，业务账号仅拥有指定数据库的表级权限。

func main() {
	// 可以不指定数据库
	dsn := "root:123@tcp(127.0.0.1:3306)/"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建一个新的，带 ctx 的 GORM 会话。
	// 它不会修改原来的 db 对象，因此可以在同一个 db 对象上创建多个带不同 ctx 的会话。
	tx := db.WithContext(ctx)

	err = tx.Exec(`
                CREATE DATABASE IF NOT EXISTS demo
                CHARACTER SET utf8mb4
        `).Error
	if err != nil {
		log.Fatal(err)
	}
	// 创建数据库后，如果要操作其中的表，通常需要重新建立带数据库名的连接
	log.Println("数据库创建成功")

	if err = tx.Exec("DROP DATABASE IF EXISTS demo").Error; err != nil {
		log.Fatal(err)
	}
	log.Println("数据库删除成功")
}
