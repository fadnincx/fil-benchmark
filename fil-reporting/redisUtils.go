package fil_reporting

import (
	testbed "fil-benchmark/external-testbed"
	"github.com/go-redis/redis"
	"sync"
)

var redisClient *redis.Client = nil
var redisMutex sync.Mutex

func init() {
	redisInitClient()
}

func redisInitClient() {
	redisMutex.Lock()
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     testbed.GetTestBed().GetRedisHost() + ":6379",
			Password: "",
			DB:       0,
		})
	}
	redisMutex.Unlock()
}
