package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// 创建和删除数据库
// 生产环境一般不建议让业务服务随意创建或删除数据库，更常见的是通过部署脚本或管理员账号执行，业务账号仅拥有指定数据库的表级权限。

func main() {
	// 可以不指定数据库
	dsn := "root:123@tcp(127.0.0.1:3306)/"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	_, err = db.ExecContext(ctx, `
                CREATE DATABASE IF NOT EXISTS demo
                CHARACTER SET utf8mb4
        `)
	if err != nil {
		log.Fatal(err)
	}
	// 创建数据库后，如果要操作其中的表，通常需要重新建立带数据库名的连接
	log.Println("数据库创建成功")

	if _, err = db.ExecContext(ctx, "DROP DATABASE IF EXISTS demo"); err != nil {
		log.Fatal(err)
	}
	log.Println("数据库删除成功")
}
