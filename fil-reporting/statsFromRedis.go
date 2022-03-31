package fil_reporting

import (
	"encoding/json"
	"fil-benchmark/datastructures"
	"fmt"
	"github.com/go-redis/redis"
)

/**
 * Get stats about processed messages
 */
func redisGetMsgStats(cids []string, hosts []string) []datastructures.Stats {

	// Make sure client is initialized
	redisInitClient()

	// Init stats per host
	var stats []datastructures.Stats

	// Get stats per host
	for _, host := range hosts {

		// Init stats
		var hostStats datastructures.Stats
		hostStats.Min = ^uint64(0)

		// Get processing time of entries origin on host
		var entries []uint64

		// Iterate over given cids
		for _, cid := range cids {

			// Get cid date on host
			val, err := redisClient.Get(cid + "-" + host).Result()
			if err == redis.Nil {
				// Ignore no such entry
			} else if err != nil {
				fmt.Println(err) // some redis error
			} else {

				// Get object from json
				var stored datastructures.TimeLogEntry
				err = json.Unmarshal([]byte(val), &stored)
				if err != nil {
					fmt.Println(err)
				}

				// If not finished yet
				if stored.End == 0 {
					// notFinished++
					continue
				}

				// Capture illegal start > end
				if stored.Start > stored.End {
					fmt.Printf("ILLEGAL Start %v > End %v \n", stored.Start, stored.End)
					continue
				}
				// Get those entries with a start time
				if stored.Start > 0 {
					entries = append(entries, uint64(stored.End-stored.Start))
				}
			}
		}

		hostStats.Amount = uint64(len(entries))
		hostStats.Min = Min(entries)
		hostStats.Max = Max(entries)
		hostStats.Avg = Mean(entries)
		hostStats.Med = Median(entries)
		stats = append(stats, hostStats)
	}

	return stats

}

/**
 * Get network delay on processed messages
 */
func redisGetMsgNetDelay(cids []string, hosts []string) []datastructures.Stats {

	// Make sure client is initialized
	redisInitClient()

	// Init stats per host
	var stats []datastructures.Stats
	var entries [][][]uint64 = make([][][]uint64, len(hosts))
	// Get stats per host
	for i, _ := range hosts {
		entries[i] = make([][]uint64, len(hosts))
		for j, _ := range hosts {
			entries[i][j] = make([]uint64, 0)
		}
	}

	// Iterate over given cids
	for _, cid := range cids {

		var minEnd = ^uint64(0)
		var minHost = 0

		for i, host := range hosts {

			// Get cid date on host
			val, err := redisClient.Get(cid + "-" + host).Result()
			if err == redis.Nil {
				// Ignore no such entry
			} else if err != nil {
				fmt.Println(err) // some redis error
			} else {

				// Get object from json
				var stored datastructures.TimeLogEntry
				err = json.Unmarshal([]byte(val), &stored)
				if err != nil {
					fmt.Println(err)
				}

				// If not finished yet
				if stored.End > 0 && uint64(stored.End) < minEnd {
					minEnd = uint64(stored.End)
					minHost = i
					break
				}
			}

		}

		if minEnd != 0 {

			for i, host := range hosts {
				// Get cid date on host
				val, err := redisClient.Get(cid + "-" + host).Result()
				if err == redis.Nil {
					// Ignore no such entry
				} else if err != nil {
					fmt.Println(err) // some redis error
				} else {

					// Get object from json
					var stored datastructures.TimeLogEntry
					err = json.Unmarshal([]byte(val), &stored)
					if err != nil {
						fmt.Println(err)
					}

					// If not finished yet
					if stored.End != 0 {
						entries[minHost][i] = append(entries[minHost][i], uint64(stored.End)-minEnd)
					}
				}
			}

		}

	}
	for i, _ := range hosts {
		for j, _ := range hosts {
			if i != j {
				// Init stats
				var hostStats datastructures.Stats
				hostStats.Min = ^uint64(0)
				hostStats.Min = Min(entries[i][j])
				hostStats.Max = Max(entries[i][j])
				hostStats.Avg = Mean(entries[i][j])
				hostStats.Med = Median(entries[i][j])
				hostStats.Amount = uint64(len(entries[i][j]))
				hostStats.Desc = fmt.Sprintf("Msg-Delay: %s -> %s", hosts[i], hosts[j])
				stats = append(stats, hostStats)
			}

		}

	}

	return stats

}
