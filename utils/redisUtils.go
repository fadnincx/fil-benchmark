package utils

import (
	testbed "fil-benchmark/external-testbed"
	"github.com/go-redis/redis"
	"sync"
)

var rhelper *RedisHelper

type RedisHelper struct {
	RedisClient *redis.Client
	RedisMutex  sync.Mutex
}

func init() {
	rhelper = &RedisHelper{
		RedisClient: nil,
	}
	rhelper.redisInitClient()
}
func GetRedisHelper() *RedisHelper {
	return rhelper
}
func (rh *RedisHelper) redisInitClient() {
	rh.RedisMutex.Lock()
	if rh.RedisClient == nil {
		rh.RedisClient = redis.NewClient(&redis.Options{
			Addr:     testbed.GetTestBed().GetRedisHost() + ":6379",
			Password: "",
			DB:       0,
		})
	}
	rh.RedisMutex.Unlock()
}
