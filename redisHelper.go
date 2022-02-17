package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"sync"
)

type TimeLogEntry struct {
	Client string `json:"client"`
	Cid    string `json:"cid"`
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
}
type DelayEntry struct {
	avgDelay float64
	amount   uint64
}

var redisClient *redis.Client = nil
var redisMutex sync.Mutex

func init() {
	redisInitClient()
}

func redisInitClient() {
	redisMutex.Lock()
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     getRedisNode() + ":6379",
			Password: "",
			DB:       0,
		})
	}
	redisMutex.Unlock()
}

func redisGetMsgWith2Times() {
	redisInitClient()
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = redisClient.Scan(cursor, "*", 0).Result()
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, key := range keys {
			val, err := redisClient.Get(key).Result()
			if err != nil {
				fmt.Println(err)
			}
			var stored TimeLogEntry
			err = json.Unmarshal([]byte(val), &stored)
			if err != nil {
				fmt.Println(err)
			}
			if stored.Start == 0 {
				continue
			}
			if stored.End == 0 {
				continue
			}
			fmt.Printf("Found %v\n", val)
		}

		if cursor == 0 { // no more keys
			break
		}
	}
}

func redisGetAvgDelay(cids []string, hosts []string) (uint64, float64, uint64, []DelayEntry) {
	redisInitClient()
	var amountOfResults uint64 = 0
	var averageDelay float64 = 0
	var notFinished uint64 = 0
	var netDelay = make([]DelayEntry, len(hosts))
	for _, cid := range cids {

		var netT = make([]int64, len(hosts))
		var netS int64 = 0
		for i, host := range hosts {
			val, err := redisClient.Get(cid + "-" + host).Result()
			if err == redis.Nil {
				// Ignore no such entry
			} else if err != nil {
				fmt.Println(err)
			} else {
				var stored TimeLogEntry
				err = json.Unmarshal([]byte(val), &stored)
				if err != nil {
					fmt.Println(err)
				}
				// fmt.Printf("Redis: %v --> %v\n", cid+"-"+host, val)

				if stored.End == 0 {
					notFinished++
					continue
				}
				if stored.Start > stored.End {
					fmt.Printf("ILLEGAL Start %v > End %v \n", stored.Start, stored.End)
					continue
				}
				if stored.Start > 0 {
					netS = stored.End
					averageDelay = ((averageDelay * float64(amountOfResults)) + float64(stored.End-stored.Start)) / float64(amountOfResults+1)
					amountOfResults += 1
					// fmt.Printf("%d = %d - %d  --> %f %d\n", stored.End-stored.Start, stored.End, stored.Start, averageDelay, amountOfResults)
				}
				netT[i] = stored.End
			}
		}

		if netS > 0 {
			for i := range hosts {
				if netT[i] != 0 {
					netDelay[i].avgDelay = ((netDelay[i].avgDelay * float64(netDelay[i].amount)) + float64(netT[i]-netS)) / float64(netDelay[i].amount+1)
					netDelay[i].amount += 1
				}
			}
		}

	}
	fmt.Printf("%v unfinished messages found\n", notFinished)
	return amountOfResults, averageDelay, notFinished, netDelay
}
func redisHasNonFinished(cids []string, hosts []string) bool {
	redisInitClient()
	for _, cid := range cids {
		for _, host := range hosts {
			val, err := redisClient.Get(cid + "-" + host).Result()
			if err == redis.Nil {
				// Ignore no such entry
			} else if err != nil {
				fmt.Println(err)
			} else {
				var stored TimeLogEntry
				err = json.Unmarshal([]byte(val), &stored)
				if err != nil {
					fmt.Println(err)
				}

				if stored.End == 0 {
					return true
				}
			}
		}
	}
	return false
}

func redisScanAverageDelay(nodeHostname string, from int64, to int64) (uint64, float64, error) {
	redisInitClient()
	var cursor uint64
	var amountOfResults uint64 = 0
	var averageDelay float64 = 0
	for {
		var keys []string
		var err error
		keys, cursor, err = redisClient.Scan(cursor, "*-"+nodeHostname+"*", 0).Result()
		if err != nil {
			fmt.Println(err)
			return 0, 0.0, errors.New("Error scanning redis")
		}
		fmt.Printf("Redis keys: %v \n", keys)
		for _, key := range keys {
			val, err := redisClient.Get(key).Result()
			if err != nil {
				fmt.Println(err)
			}
			var stored TimeLogEntry
			err = json.Unmarshal([]byte(val), &stored)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("Redis: %v\n", val)
			if stored.Start < from {
				continue
			}
			if stored.Start > to {
				continue
			}
			if stored.End == 0 {
				continue
			}
			averageDelay = ((averageDelay * float64(amountOfResults)) + float64(stored.End-stored.Start)) / float64(amountOfResults+1)
			amountOfResults += 1
		}

		if cursor == 0 { // no more keys
			break
		}
	}

	return amountOfResults, averageDelay, nil

}
