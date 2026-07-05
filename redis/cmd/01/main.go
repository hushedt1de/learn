package main

import (
	"context"
	"log"
	"redis-learn/utils/redis"
)

func main() {
	// 连接 redis
	res, err := redis.GetClient().Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
	log.Println(res)
}
