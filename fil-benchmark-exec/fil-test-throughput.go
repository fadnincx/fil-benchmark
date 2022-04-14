package fil_benchmark_exec

import (
	"encoding/json"
	"fil-benchmark/datastructures"
	"fil-benchmark/fil-reporting"
	"fmt"
	"log"
	"strconv"
	"time"
)

func RunTestcases(cases []datastructures.TestCase, nodes []datastructures.Node, report fil_reporting.ReportResults) {

	// Setup all websocket connections to the nodes
	var wsWriteChan = make([]chan string, len(nodes))
	var wsReadChan = make([]chan string, len(nodes))
	var wsInteruptChan = make([]chan bool, len(nodes))
	for i := range nodes {

		wsWriteChan[i] = make(chan string, 2) // Size 2 as a small buffer, when on limit to channel saturation
		wsReadChan[i] = make(chan string, 2)  // Size 2 as a small buffer, when on limit to channel saturation

		// Process Send response and output CIDs to reporting
		go func(input chan string, cidChan chan string) {
			for {
				s := <-input
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
		}(wsReadChan[i], report.CidChan)
		wsInteruptChan[i] = make(chan bool)

		// Start Websocket connection
		go connect2LotusWsApi(nodes[i], wsInteruptChan[i], wsWriteChan[i], wsReadChan[i])
	}

	for _, testc := range cases {

		log.Printf("Start test case %v\n", testc)

		// Wait unitl ready for next
		log.Println("Wait for logging to be ready")
		<-report.ReadyForNextTest

		var stopRateChan = make([]chan bool, len(nodes))

		report.TestStatusChan <- "startrate"
		report.TestDataChain <- fmt.Sprintf("%v", testc.MsgPerSec)

		// Start rate
		for i := range nodes {
			stopRateChan[i] = make(chan bool)
			go sendAtRate(wsWriteChan[i], stopRateChan[i], testc.MsgPerSec/float64(len(nodes)), nodes[i].SendWallet, nodes[(i+1)%len(nodes)].SendWallet, "10")
		}

		report.TestStatusChan <- "waitsaturation"

		// Wait 90s for full saturation
		time.Sleep(90 * time.Second)

		report.TestStatusChan <- "start"

		// Sleep duration
		time.Sleep(time.Duration(testc.Duration * 1e9))

		report.TestStatusChan <- "stop"

		// Stop rate
		for i := range nodes {
			stopRateChan[i] <- true
		}

	}
	log.Println("Finished tests")

	// Stop websocket
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