package main

import (
	"context"
	"log"
	"redis-learn/utils"
)

func main() {
	// 连接 redis
	res, err := utils.GetClient().Ping(context.Background()).Result()
	if err != nil {
		log.Panic(err)
	}
	log.Println(res)
}
