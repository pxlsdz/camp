package goCache

import (
	"github.com/coocood/freecache"
	"runtime/debug"
)

var cache *freecache.Cache

func Init() {
	// 缓存大小，200M
	cacheSize := 200 * 1024 * 1024
	cache = freecache.NewCache(cacheSize)
	debug.SetGCPercent(20)
	//key := []byte("abc")
	//val := []byte("def")
	//expire := 60 // expire in 60 seconds
	//cache.Set(key, val, expire)
	//got, err := cache.Get(key)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%s\n", got)
	//}
	//affected := cache.Del(key)
	//fmt.Println("deleted key ", affected)
	//fmt.Println("entry count ", cache.EntryCount())
}

func GetCache() *freecache.Cache {
	return cache
}
