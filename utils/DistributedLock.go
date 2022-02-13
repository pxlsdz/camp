package utils

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type DistributedLock struct {
	cli        *redis.Client
	ctx        context.Context
	lockKey    string
	expiration time.Duration
	unlockCh   chan struct{}
}

func NewDistributedLock(cli *redis.Client, ctx context.Context, lockKey string, expiration time.Duration) *DistributedLock {
	return &DistributedLock{
		cli:        cli,
		ctx:        ctx,
		lockKey:    lockKey,
		expiration: expiration,
		unlockCh:   make(chan struct{}, 0),
	}
}

func (dl DistributedLock) Lock() {
	var resp *redis.BoolCmd
	for {
		goId := GetCurrentGoroutineId()
		resp = dl.cli.SetNX(dl.ctx, dl.lockKey, goId, dl.expiration) //返回执行结果

		lockSuccess, err := resp.Result()
		if err == nil && lockSuccess {
			//抢锁成功，开启看门狗 并跳出，否则失败继续自旋
			go dl.watchDog(goId)
			return
		} else {
			//time.Sleep(time.Millisecond*30) //可以适当休眠
		}

	}

}

//自动续期看门狗

func (dl DistributedLock) watchDog(goId int) {

	// 创建一个定时器NewTicker, 每隔8秒触发一次,类似于闹钟
	expTicker := time.NewTicker(time.Second * 8)

	//确认锁与锁续期打包原子化

	script := redis.NewScript(`
    if redis.call('get', KEYS[1]) == ARGV[1]
    then
      return redis.call('expire', KEYS[1], ARGV[2])
    else
      return 0
    end
  `)

	for {
		select {
		case <-expTicker.C: //因为上边是用NewTicker创建的定时器，所以每隔8s都会触发

			resp := script.Run(dl.ctx, dl.cli, []string{dl.lockKey}, goId, 10)
			if result, err := resp.Result(); err != nil || result == int64(0) {
				//续期失败
				log.Println("expire lock failed", err)

			}
		case <-dl.unlockCh: //任务完成后用户解锁通知看门狗退出
			return
		}

	}

}

func (dl DistributedLock) Unlock() error {

	script := redis.NewScript(`
    if redis.call('get', KEYS[1]) == ARGV[1]
    then
      return redis.call('del', KEYS[1])
     else
        return 0
     end

  `)

	resp := script.Run(dl.ctx, dl.cli, []string{dl.lockKey}, GetCurrentGoroutineId())

	if result, err := resp.Result(); err != nil || result == 0 {
		return err

	} else {
		//删锁成功后，通知看门狗退出
		dl.unlockCh <- struct{}{}

	}
	return nil

}

func GetCurrentGoroutineId() int {

	buf := make([]byte, 128)

	buf = buf[:runtime.Stack(buf, false)]

	stackInfo := string(buf)

	goIdStr := strings.TrimSpace(strings.Split(strings.Split(stackInfo, "[running]")[0], "goroutine")[1])
	goId, err := strconv.Atoi(goIdStr)

	if err != nil {
		fmt.Println("err=", err)
		return 0

	}

	return goId

}
