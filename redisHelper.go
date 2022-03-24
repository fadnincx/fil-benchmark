package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"sort"
	"sync"
)

type TimeLogEntry struct {
	Client string `json:"client"`
	Cid    string `json:"cid"`
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
}
type BlockLog struct {
	Client   string `json:"client"`
	Cid      string `json:"cid"`
	MsgCount uint64 `json:"amount"`
	Time     int64  `json:"time"`
}
type Stats struct {
	min    uint64
	max    uint64
	avg    uint64
	med    uint64
	amount uint64
	desc   string
}
type BlockStats struct {
	minBlockInterval    uint64
	maxBlockInterval    uint64
	avgBlockInterval    uint64
	medianBlockInterval uint64
	minMsgPerBlock      uint64
	maxMsgPerBlock      uint64
	avgMsgPerBlock      uint64
	medianMsgPerBlock   uint64
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

/**
 * Get stats about processed messages
 */
func redisGetMsgStats(cids []string, hosts []string) []Stats {

	// Make sure client is initialized
	redisInitClient()

	// Init stats per host
	var stats []Stats = make([]Stats, len(hosts))

	// Get stats per host
	for _, host := range hosts {

		// Init stats
		var hostStats Stats
		hostStats.min = ^uint64(0)

		// Get processing time of entries origin on host
		var entries = make([]uint64, len(cids))

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
				var stored TimeLogEntry
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
		hostStats.min = Min(entries)
		hostStats.max = Max(entries)
		hostStats.avg = Mean(entries)
		hostStats.med = Median(entries)
		hostStats.amount = uint64(len(entries))
		stats = append(stats, hostStats)
	}

	return stats

}

/**
 * Get network delay on processed messages
 */
func redisGetMsgNetDelay(cids []string, hosts []string) []Stats {

	// Make sure client is initialized
	redisInitClient()

	// Init stats per host
	var stats []Stats = make([]Stats, len(hosts))
	var entries [][][]uint64 = make([][][]uint64, len(hosts))
	// Get stats per host
	for i, _ := range hosts {
		entries[i] = make([][]uint64, len(hosts))
		for j, _ := range hosts {
			entries[i][j] = make([]uint64, len(cids))
		}
	}

	// Iterate over given cids
	for _, cid := range cids {

		var minEnd uint64 = ^uint64(0)
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
				var stored TimeLogEntry
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
					var stored TimeLogEntry
					err = json.Unmarshal([]byte(val), &stored)
					if err != nil {
						fmt.Println(err)
					}

					// If not finished yet
					if stored.End != 0 {
						entries[i][minHost] = append(entries[i][minHost], uint64(stored.End)-minEnd)
					}
				}
			}

		}

	}
	for i, _ := range hosts {
		for j, _ := range hosts {
			if i != j {

				// Init stats
				var hostStats Stats
				hostStats.min = ^uint64(0)
				hostStats.min = Min(entries[i][j])
				hostStats.max = Max(entries[i][j])
				hostStats.avg = Mean(entries[i][j])
				hostStats.med = Median(entries[i][j])
				hostStats.amount = uint64(len(entries[i][j]))
				hostStats.desc = fmt.Sprintf("Msg-Delay: %s -> %s", hosts[j], hosts[i])
				stats = append(stats, hostStats)
			}

		}

	}

	return stats

}

/**
 *
 */

func chainBlockAnalysis(blocks []BlockLog) BlockStats {

	var stats BlockStats

	stats.minBlockInterval = ^uint64(0)
	stats.minMsgPerBlock = ^uint64(0)

	blockIntervals := make([]uint64, len(blocks)-1)

	for i, b := range blocks {

		if i > 0 {
			interval := uint64(blocks[i].Time - blocks[i-1].Time)

			if interval < stats.minBlockInterval {
				stats.minBlockInterval = interval
			}
			if interval > stats.maxBlockInterval {
				stats.maxBlockInterval = interval
			}

			blockIntervals = append(blockIntervals, interval)
		}
		if b.MsgCount < stats.minMsgPerBlock {
			stats.minMsgPerBlock = b.MsgCount
		}
		if b.MsgCount > stats.maxMsgPerBlock {
			stats.maxMsgPerBlock = b.MsgCount
		}

	}

	msgs := Map(blocks, func(b BlockLog) uint64 { return b.MsgCount })

	stats.minBlockInterval = Min(blockIntervals)
	stats.maxBlockInterval = Max(blockIntervals)
	stats.avgBlockInterval = Mean(blockIntervals)
	stats.medianBlockInterval = Median(blockIntervals)
	stats.minMsgPerBlock = Min(msgs)
	stats.maxMsgPerBlock = Max(msgs)
	stats.avgMsgPerBlock = Mean(msgs)
	stats.medianMsgPerBlock = Median(msgs)

	return stats

}
func redisGetBlocks(from int64, to int64, node string) []BlockLog {
	redisInitClient()
	var cursor uint64
	var output []BlockLog = make([]BlockLog, 0)
	for {
		var keys []string
		var err error
		keys, cursor, err = redisClient.Scan(cursor, "*-p-"+node+"*", 0).Result()
		if err != nil {
			fmt.Println(err)
			return nil
		}
		for _, key := range keys {
			val, err := redisClient.Get(key).Result()
			if err == redis.Nil {
				// Ignore no such entry
			} else if err != nil {
				fmt.Println(err)
			} else {
				var stored BlockLog
				err = json.Unmarshal([]byte(val), &stored)
				if err != nil {
					fmt.Println(err)
				}
				if stored.Time >= from && stored.Time <= to {
					output = append(output, stored)
				}
			}
		}

		if cursor == 0 { // no more keys
			break
		}
	}
	sort.SliceStable(output, func(i, j int) bool {
		return output[i].Time < output[j].Time
	})
	return output
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
