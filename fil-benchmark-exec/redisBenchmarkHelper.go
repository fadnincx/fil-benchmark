package fil_benchmark_exec

import (
	"encoding/json"
	"errors"
	"fil-benchmark/datastructures"
	"fil-benchmark/utils"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

func returnWhenMessageIsAccepted(cid string) error {

	for {

		var cursor uint64
		for {
			var keys []string
			var err error
			keys, cursor, err = utils.GetRedisHelper().GetClient().Scan(cursor, cid+"-*", 0).Result()
			if err != nil {
				fmt.Println(err)
				return errors.New("error scanning redis")
			}
			for _, key := range keys {
				val, err := utils.GetRedisHelper().GetClient().Get(key).Result()
				if err == redis.Nil {
					// Ignore no such entry
				} else if err != nil {
					fmt.Println(err)
				} else {
					var stored datastructures.TimeLogEntry
					err = json.Unmarshal([]byte(val), &stored)
					if err != nil {
						fmt.Println(err)
					}
					if stored.End > 0 {
						return nil
					}
				}
			}

			if cursor == 0 { // no more keys
				break
			}

		}

		time.Sleep(15 * time.Second)

	}

}
