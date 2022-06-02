package fil_reporting

import (
	fil_benchmark_exec "fil-benchmark/datastructures"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type ReportResults struct {
	CidChan          chan string
	TestStatusChan   chan string
	TestDataChain    chan string
	ReadyForNextTest chan bool
}

func StartTestReport(nodes []fil_benchmark_exec.Node) ReportResults {

	results := ReportResults{
		CidChan:          make(chan string),
		TestStatusChan:   make(chan string),
		TestDataChain:    make(chan string),
		ReadyForNextTest: make(chan bool, 1),
	}

	go testReport(nodes, results.CidChan, results.TestStatusChan, results.TestDataChain, results.ReadyForNextTest)

	return results
}
func testReport(nodes []fil_benchmark_exec.Node, cids chan string, testStatus chan string, testDataChain chan string, ready4Next chan bool) {

	var hosts []string
	for _, n := range nodes {
		hosts = append(hosts, n.Hostname)
	}

	// Capture all Cids together
	cidDo := make(chan string)
	cidAgg := make(chan []string)

	go func(cid chan string, do chan string, out chan []string) {
		agg := make([]string, 0, 10000)
		for {
			select {
			case c := <-cid:
				agg = append(agg, c)
			case d := <-do:
				switch d {
				case "get":
					out <- agg
					agg = make([]string, 0, 10000)
				case "drop":
					agg = make([]string, 0, 10000)
				}

			}
		}
	}(cids, cidDo, cidAgg)

	var starttime int64
	var stoptime int64
	var rate float64
	ready4Next <- true
	log.Println("Log ready to start")
	for {
		select {
		case d := <-testStatus:
			switch d {
			case "startrate":
				rate, _ = strconv.ParseFloat(<-testDataChain, 64)
				log.Println("Start with load on network")
			case "waitsaturation":
				log.Println("Wait until network is saturated")
			case "start":
				starttime = time.Now().UnixMicro()
				cidDo <- "drop"
				log.Println("Start test")
			case "stop":
				stoptime = time.Now().UnixMicro()
				cidDo <- "get"
				go doStats(hosts, rate, <-cidAgg, uint64(starttime), uint64(stoptime), ready4Next)

			default:
				log.Printf("Warning unknown test status receiverd: '%s'\n", d)
			}
		}
	}
}
func doStats(hosts []string, rate float64, cids []string, starttime uint64, stoptime uint64, ready4Next chan bool) {

	// Get Msg Stats
	msgStats := redisGetMsgStats(cids, hosts)
	var msgStat []string
	var msgStatHeader []string
	for i, v := range msgStats {
		msgStat = append(msgStat, fmt.Sprintf("%d, %d, %d, %d, %d", v.Amount, v.Min, v.Max, v.Avg, v.Med))
		msgStatHeader = append(msgStatHeader, fmt.Sprintf("amount%d, min%d, max%d, avg%d, med%d", i, i, i, i, i))
	}

	// Get Msg Network Delay
	/*	msgNetDelays := redisGetMsgNetDelay(cids, hosts)
		var msgNetDelay []string
		var msgNetDelayHeader []string
		for _, v := range msgNetDelays {
			msgNetDelay = append(msgNetDelay, fmt.Sprintf("%v, %d, %d, %d, %d, %d", v.Desc, v.Amount, v.Min, v.Max, v.Avg, v.Med))
			msgNetDelayHeader = append(msgNetDelayHeader, "desc, amount, min, max, avg, med")
		}*/

	// Get Multiple Block stats
	multiBlockStats := redisGetMsgInMultipleBlocksStats(hosts[0], int64(starttime), int64(stoptime))
	multiBlockStats2 := redisGetMsgInMultipleTipsets(hosts[0], int64(starttime), int64(stoptime))
	var multiBlockStat []string
	var multiBlockStatHeader []string
	for i := range multiBlockStats {
		multiBlockStat = append(multiBlockStat, fmt.Sprintf("%d, %d", multiBlockStats[i], multiBlockStats2[i]))
		multiBlockStatHeader = append(multiBlockStatHeader, fmt.Sprintf("# in %d block, # in %d tipsets", i, i))
	}

	line := fmt.Sprintf("%f, %v, %v, %v\n", rate, strings.Join(msgStat, ", "), strings.Join(multiBlockStat, ", "))
	header := fmt.Sprintf("rate, %v, %v, %v\n", strings.Join(msgStatHeader, ", "), strings.Join(multiBlockStatHeader, ", "))

	timeLogCsv.writeCsvLine(line, header)
	fmt.Printf("Rate %v has %v messages with %vus delay\n", rate, msgStats[0].Amount, msgStats[0].Avg)

	doBlockStats(hosts, int64(starttime), int64(stoptime))
	ready4Next <- true
}

func doBlockStats(hosts []string, starttime int64, stoptime int64) {
	aggBlockStats := redisGetBlockStats(hosts, starttime, stoptime)
	var headerBlocks []string
	headerBlocks = append(headerBlocks, fmt.Sprintf("cid, minFirstSeen, minFirstSeenHost"))
	for i, _ := range hosts {
		headerBlocks = append(headerBlocks, fmt.Sprintf("firstSeen%d, approved%d", i, i))
	}
	header := strings.Join(headerBlocks, ", ")

	for _, b := range aggBlockStats {
		var lineBlocks []string
		lineBlocks = append(lineBlocks, fmt.Sprintf("%s, %d, %d", b.Cid, b.MinTime, b.MinTimeHost))
		for i, _ := range hosts {
			lineBlocks = append(lineBlocks, fmt.Sprintf("%d, %d", b.FirstKnown[i], b.Accepted[i]))
		}
		blockLogCsv.writeCsvLine(strings.Join(lineBlocks, ", "), header)

	}

}
