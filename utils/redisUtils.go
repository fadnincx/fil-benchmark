package utils

import (
	testbed "fil-benchmark/external-testbed"
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"strings"
	"sync"
)

var rhelper *RedisHelper

type RedisHelper struct {
	redisClient *redis.Client
	redisMutex  sync.Mutex
}

func init() {
	rhelper = &RedisHelper{
		redisClient: nil,
	}
	rhelper.redisInitClient()
}
func GetRedisHelper() *RedisHelper {
	return rhelper
}
func (rh *RedisHelper) GetClient() *redis.Client {
	rh.redisMutex.Lock()
	_, err := rh.redisClient.Ping().Result()
	if err != nil {
		rh.redisClient = nil
		rh.redisInitClient()
	}
	return rh.redisClient
}
func (rh *RedisHelper) redisInitClient() {
	rh.redisMutex.Lock()
	if rh.redisClient == nil {
		redisIP := testbed.GetTestBed().GetRedisHost()
		if !is_ipv4(redisIP) {
			log.Printf("Redis IP %s is not valid ipv4", redisIP)
			rh.redisMutex.Unlock()
			return
		}
		rh.redisClient = redis.NewClient(&redis.Options{
			Addr:     redisIP + ":6379",
			Password: "",
			DB:       0,
		})
		log.Printf("Redis initialized with ip %s\n", redisIP)
	}
	rh.redisMutex.Unlock()
}
func is_ipv4(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) < 4 {
		return false
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return false
			}
		} else {
			return false
		}

	}
	return true
}
