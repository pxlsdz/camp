package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"time"
)

const Nil = redis.Nil

var cli *redis.Client

var LuaHash string

var script string = `
		local shc = redis.call("Get", KEYS[1])
		if (shc) then
			return 3
		end

		local cap = redis.call('Get', KEYS[2])
		if (not cap) then
			return 2
		elseif (tonumber(cap) > 0)  then
			redis.call("DecrBy" , KEYS[2], 1)
			redis.call("Set" , KEYS[1], 1)
			return 1
		else
			return 0
		end
		return -1
	`

func Init() error {
	addr := viper.GetString("redis.addr")
	cli = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	LuaHash, _ = cli.ScriptLoad(ctx, script).Result()
	fmt.Println(LuaHash)
	return cli.Ping(ctx).Err()
}

func GetClient() *redis.Client {
	return cli
}
