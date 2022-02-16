package myRedis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"time"
)

const Nil = redis.Nil

var cli *redis.Client

var LuaHash string

var script string = `
		local s = redis.call("Get", KEYS[3])
		if (not s) then
			return 4
		end

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
	password := viper.GetString("redis.password")

	cli = redis.NewClient(&redis.Options{
		//Network:      "unix",
		Addr:         addr,
		Password:     password,
		DB:           0,
		MinIdleConns: 16,                //空闲连接数
		MaxConnAge:   300 * time.Second, //空闲5分钟后关闭连接
		PoolSize:     1000,              //最大连接数
	})
	//cli = redis.NewClient(&redis.Options{
	//	Addr:     addr,
	//	Password: password,
	//	DB:       0,
	//})
	//cli = redis.NewClient(&redis.Options{
	//	Network:  "unix",
	//	Addr:     addr,
	//	Password: password,
	//	DB:       0,
	//})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// LuaHash, _ = cli.ScriptLoad(ctx, script).Result()
	return cli.Ping(ctx).Err()
}

func GetClient() *redis.Client {
	return cli
}
