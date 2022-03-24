package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TestCase struct {
	msgPerSec float64
	duration  uint64
}

func throughputTest(nodes []node, hosts []string, cases []TestCase) {

	go outputCSVHeader(hosts)

	cidSend := make(chan string)
	cidAgg := make(chan []string)
	cidDoAgg := make(chan bool)

	var wsWriteChan = make([]chan string, len(hosts))
	var wsReadChan = make([]chan string, len(hosts))
	var wsInteruptChan = make([]chan bool, len(hosts))
	for i := range nodes {

		wsWriteChan[i] = make(chan string, 2)
		go monitorChanOverflow(wsWriteChan[i], func() { fmt.Println("WARNING WS SEND CHAN OVERFLOW!") })
		wsReadChan[i] = make(chan string, 2)
		go monitorChanOverflow(wsReadChan[i], func() { fmt.Println("WARNING WS SEND CHAN OVERFLOW!") })
		// Process Send response and output CIDs
		go func(input chan string, cidChan chan string) {
			for {
				s := <-input
				// fmt.Println(s)
				var f interface{}
				json.Unmarshal([]byte(s), &f)

				m := f.(map[string]interface{})
				resultmap := m["result"]
				r := resultmap.(map[string]interface{})
				cidmap := r["CID"]
				c := cidmap.(map[string]interface{})
				cid := c["/"].(string)

				cidChan <- cid

			}
		}(wsReadChan[i], cidSend)
		wsInteruptChan[i] = make(chan bool)
	}
	// aggregate incoming cids until next output/Flush
	go func(in chan string, out chan []string, sendOut chan bool) {
		cids := make([]string, 100)
		for {

			select {
			case c := <-in:
				cids = append(cids, c)
			case <-sendOut:
				out <- cids
				cids = make([]string, 100)
			}

		}
	}(cidSend, cidAgg, cidDoAgg)

	// Start Websocket connection
	for i := range nodes {
		go websocketC(nodes[i].ip, wsInteruptChan[i], wsWriteChan[i], wsReadChan[i])
	}

	for _, testc := range cases {

		var stopRateChan = make([]chan bool, len(hosts))

		// Start rate
		for i := range nodes {
			stopRateChan[i] = make(chan bool)
			go sendAtRate(wsWriteChan[i], stopRateChan[i], testc.msgPerSec/float64(len(nodes)), nodes[i].wallet, nodes[(i+1)%len(nodes)].wallet, "10")
		}

		// Wait 90s for full saturation
		time.Sleep(90 * time.Second)

		measureStartTime := time.Now().UnixMicro()
		// Clear cids
		cidDoAgg <- true
		_ = <-cidAgg

		// Sleep duration
		time.Sleep(time.Duration(testc.duration * 1e9))

		// Stop rate
		for i := range nodes {
			stopRateChan[i] <- true
		}

		measureStopTime := time.Now().UnixMicro()
		cidDoAgg <- true
		cids := <-cidAgg

		go outputCSVLine(cids, hosts, testc.msgPerSec, measureStartTime, measureStopTime)

		// Wait until all messages are done
		for redisHasNonFinished(cids, hosts) {
			time.Sleep(5 * time.Second)
		}

	}

	for i := range nodes {
		wsInteruptChan[i] <- true
	}

}
func sendAtRate(sendChannel chan string, stop chan bool, rate float64, senderAddress string, receiverAddress string, amount string) {
	fmt.Printf("Set rate to %v\n", rate)

	lastTime := time.Now().UnixNano()

	for sendId := 5; ; sendId++ {
		select {
		case <-stop:
			return
		default:
			sendChannel <- "{\"jsonrpc\":\"2.0\",\"method\":\"Filecoin.MpoolPushMessage\",\"params\":[{\"Version\":0,\"To\":\"" + receiverAddress + "\",\"From\":\"" + senderAddress + "\",\"Value\":\"" + amount + "\",\"GasLimit\":0,\"GasFeeCap\":\"0\",\"GasPremium\":\"0\",\"Method\":0,\"CID\":{}},{\"MaxFee\":\"0\"}],\"id\":" + strconv.Itoa(sendId) + "}"

			time.Sleep(time.Duration(int64(float64(1/rate)*1e9) - (time.Now().UnixNano() - lastTime)))
			lastTime = time.Now().UnixNano()
		}

	}

}
func outputCSVHeader(hosts []string) {
	var msgStat []string
	for i := range hosts {
		msgStat = append(msgStat, fmt.Sprintf("amount%d, min%d, max%d, avg%d, med%d", i, i, i, i, i))
	}
	var msgNetDelay []string
	for i := 0; i < len(hosts)*len(hosts)-len(hosts); i++ {
		maxHost := len(hosts) - 1
		base := i / (maxHost)
		jBase := i - (base * maxHost)
		if jBase >= base {
			jBase++
		}
		comb := fmt.Sprintf("%d -> %d", base, jBase)
		msgNetDelay = append(msgNetDelay, fmt.Sprintf("desc %v, amount %v, min %v, max %v, avg %v, med %v", comb, comb, comb, comb, comb, comb))
	}
	writeCsvLine(fmt.Sprintf("rate,%s,%s", strings.Join(msgStat, ", "), strings.Join(msgNetDelay, ", ")))
}
func outputCSVLine(cids []string, hosts []string, rate float64, start int64, stop int64) {
	fmt.Printf("%d cids to check ", len(cids))
	msgStats := redisGetMsgStats(cids, hosts)
	fmt.Println(msgStats)
	var msgStat []string
	for _, v := range msgStats {
		msgStat = append(msgStat, fmt.Sprintf("%d, %d, %d, %d, %d", v.amount, v.min, v.max, v.avg, v.med))
	}
	msgNetDelays := redisGetMsgNetDelay(cids, hosts)
	var msgNetDelay []string
	for _, v := range msgNetDelays {
		msgNetDelay = append(msgNetDelay, fmt.Sprintf("%v, %d, %d, %d, %d, %d", v.desc, v.amount, v.min, v.max, v.avg, v.med))
	}

	writeCsvLine(fmt.Sprintf("%f, %v, %v\n", rate, strings.Join(msgStat, ", "), strings.Join(msgNetDelay, ", ")))
	fmt.Printf("Rate %v has %v messages with %vus delay\n", rate, msgStats[0].amount, msgStats[0].avg)

}
