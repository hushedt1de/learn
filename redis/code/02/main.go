package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"redis-learn/code/utils"
	"time"

	"github.com/redis/go-redis/v9"
)

// redis string 实战：文章阅读量计数器

func main() {
	ctx := context.Background()

	cnt, err := GetArticleViewCount(ctx, 1)
	if err != nil {
		log.Panic(err)
	}
	log.Println(cnt)

	if err = IncrementArticleViewCount(ctx, 1); err != nil {
		log.Panic(err)
	}

	cnt, err = GetArticleViewCount(ctx, 1)
	if err != nil {
		log.Panic(err)
	}
	log.Println(cnt)
}

func IncrementArticleViewCount(ctx context.Context, articleID int64) error {
	// 1. 获取 redis 客户端
	client := utils.GetClient()

	// 2. 构建 key
	key := fmt.Sprintf("article:view:%d", articleID)

	// 3. 执行自增操作
	value, err := client.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if value == 1 {
		// key 此前不存在，设置过期时间
		if err = client.Expire(ctx, key, time.Minute*1).Err(); err != nil {
			// 需要进行回退，否则会留下一个永不过期的 key
			if _, err = client.Decr(ctx, key).Result(); err != nil {
				return fmt.Errorf("回退失败: %v, 原始错误: %v", err, err)
			}
			return err
		}
	}

	return nil
}

func GetArticleViewCount(ctx context.Context, articleID int64) (int64, error) {
	// 1. 获取 redis 客户端
	client := utils.GetClient()

	// 2. 构建 key
	key := fmt.Sprintf("article:view:%d", articleID)

	// 3. 获取阅读量
	value, err := client.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// key 不存在
			return 0, nil
		}
		return 0, err
	}

	return value, nil
}
