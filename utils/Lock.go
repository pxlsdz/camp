package utils

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"time"
)

// 重试次数
var retryTimes = 5

// 重试频率
var retryInterval = time.Millisecond * 50

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

// 锁的默认过期时间
var expiration time.Duration

// 模拟分布式业务加锁场景
func MockTest(tag string) {
	var ctx, cancel = context.WithCancel(context.Background())

	defer func() {
		// 停止goroutine
		cancel()
	}()

	// 随机value
	lockV := getRandValue()

	lockK := "EXAMPLE_LOCK"

	// 默认过期时间
	expiration = time.Millisecond * 200

	fmt.Println(tag + "尝试加锁")

	set, err := rdb.SetNX(ctx, lockK, lockV, expiration).Result()

	if err != nil {
		panic(err.Error())
	}

	// 加锁失败,重试
	if set == false && retry(ctx, rdb, lockK, lockV, expiration, tag) == false {
		fmt.Println(tag + " server unavailable, try again later")
		return
	}

	fmt.Println(tag + "成功加锁")

	// 加锁成功,新增守护线程
	go watchDog(ctx, rdb, lockK, expiration, tag)

	// 处理业务(通过随机时间延迟模拟)
	fmt.Println(tag + "等待业务处理完成...")
	time.Sleep(getRandDuration())

	// 业务处理完成
	// 释放锁
	val := delByKeyWhenValueEquals(ctx, rdb, lockK, lockV)
	fmt.Println(tag+"释放结果:", val)
}

// 释放锁
func delByKeyWhenValueEquals(ctx context.Context, rdb *redis.Client, key string, value interface{}) bool {
	lua := `
-- 如果当前值与锁值一致,删除key
if redis.call('GET', KEYS[1]) == ARGV[1] then
	return redis.call('DEL', KEYS[1])
else
	return 0
end
`
	scriptKeys := []string{key}

	val, err := rdb.Eval(ctx, lua, scriptKeys, value).Result()
	if err != nil {
		panic(err.Error())
	}

	return val == int64(1)
}

// 生成随机时间
func getRandDuration() time.Duration {
	rand.Seed(time.Now().UnixNano())
	min := 50
	max := 100
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}

// 生成随机值
func getRandValue() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

// 守护线程
func watchDog(ctx context.Context, rdb *redis.Client, key string, expiration time.Duration, tag string) {
	for {
		select {
		// 业务完成
		case <-ctx.Done():
			fmt.Printf("%s任务完成,关闭%s的自动续期\n", tag, key)
			return
			// 业务未完成
		default:
			// 自动续期
			rdb.PExpire(ctx, key, expiration)
			// 继续等待
			time.Sleep(expiration / 2)
		}
	}
}

// 重试
func retry(ctx context.Context, rdb *redis.Client, key string, value interface{}, expiration time.Duration, tag string) bool {
	i := 1
	for i <= retryTimes {
		fmt.Printf(tag+"第%d次尝试加锁中...\n", i)
		set, err := rdb.SetNX(ctx, key, value, expiration).Result()

		if err != nil {
			panic(err.Error())
		}

		if set == true {
			return true
		}

		time.Sleep(retryInterval)
		i++
	}
	return false
}
