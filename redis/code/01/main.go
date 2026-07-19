package main

import (
	"context"
	"log"
	"redis-learn/code/utils"
)

// 连接 redis

func main() {
	res, err := utils.GetClient().Ping(context.Background()).Result()
	if err != nil {
		log.Panic(err)
	}
	log.Println(res)
}
