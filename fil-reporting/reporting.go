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
				go doStats(hosts, rate, <-cidAgg, uint64(starttime), uint64(stoptime))
				ready4Next <- true
			default:
				log.Printf("Warning unknown test status receiverd: '%s'\n", d)
			}
		}
	}
}
func doStats(hosts []string, rate float64, cids []string, starttime uint64, stoptime uint64) {

	// Get Msg Stats
	msgStats := redisGetMsgStats(cids, hosts)
	var msgStat []string
	var msgStatHeader []string
	for i, v := range msgStats {
		msgStat = append(msgStat, fmt.Sprintf("%d, %d, %d, %d, %d", v.Amount, v.Min, v.Max, v.Avg, v.Med))
		msgStatHeader = append(msgStatHeader, fmt.Sprintf("amount%d, min%d, max%d, avg%d, med%d", i, i, i, i, i))
	}

	// Get Msg Network Delay
	msgNetDelays := redisGetMsgNetDelay(cids, hosts)
	var msgNetDelay []string
	var msgNetDelayHeader []string
	for _, v := range msgNetDelays {
		msgNetDelay = append(msgNetDelay, fmt.Sprintf("%v, %d, %d, %d, %d, %d", v.Desc, v.Amount, v.Min, v.Max, v.Avg, v.Med))
		msgNetDelayHeader = append(msgNetDelayHeader, "desc, amount, min, max, avg, med")
	}

	line := fmt.Sprintf("%f, %v, %v\n", rate, strings.Join(msgStat, ", "), strings.Join(msgNetDelay, ", "))
	header := fmt.Sprintf("rate, %v, %v\n", strings.Join(msgStatHeader, ", "), strings.Join(msgNetDelayHeader, ", "))

	writeCsvLine(line, header)
	fmt.Printf("Rate %v has %v messages with %vus delay\n", rate, msgStats[0].Amount, msgStats[0].Avg)
}
