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

	inputChan := make(chan string, 1)
	outputChan := make(chan string, 1)
	interuptChan := make(chan bool, 1)

	cidSend := make(chan string)
	cidAgg := make(chan []string)
	cidDoAgg := make(chan bool)

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

			// fmt.Printf("Msg: %v\n", cid)
			cidChan <- cid

		}
	}(outputChan, cidSend)

	go func(ch chan string) {
		if len(ch) == cap(ch) {
			fmt.Println("WARNING Send Channel full!")
		} else {
			time.Sleep(1)
		}
	}(inputChan)

	// Start Websocket connection
	go websocketC(nodes[0].ip, interuptChan, inputChan, outputChan)

	for _, testc := range cases {
		stopRateChan := make(chan bool)

		// Start rate
		go sendAtRate(inputChan, stopRateChan, testc.msgPerSec, nodes[0].wallet, nodes[1].wallet, "10")

		// Wait 90s for full saturation
		time.Sleep(90 * time.Second)

		// Clear cids
		cidDoAgg <- true
		_ = <-cidAgg

		// Sleep duration
		time.Sleep(time.Duration(testc.duration * 1e9))

		// Stop rate
		stopRateChan <- true

		cidDoAgg <- true
		cids := <-cidAgg

		go func(cids []string, hosts []string, rate float64) {
			amount, avg, unfinished, netDelay := redisGetAvgDelay(cids, hosts)
			netDelayS := make([]string, len(netDelay))

			for _, nDel := range netDelay {
				netDelayS = append(netDelayS, fmt.Sprintf("%f, %d", nDel.avgDelay, nDel.amount))
			}
			writeCsvLine(fmt.Sprintf("%f, %d, %f, %d, %s\n", rate, amount, avg, unfinished, strings.Join(netDelayS, ", ")))
			fmt.Printf("Rate %v has %v messages with %vus delay\n", rate, amount, avg)

		}(cids, hosts, testc.msgPerSec)

		// Wait until all messages are done
		for redisHasNonFinished(cids, hosts) {
			time.Sleep(5 * time.Second)
		}

	}

	interuptChan <- true

}
func sendAtRate(sendChannel chan string, stop chan bool, rate float64, senderAddress string, receiverAddress string, amount string) {
	fmt.Printf("Set rate to %v\n", rate)

	for sendId := 5; ; sendId++ {
		select {
		case <-stop:
			return
		default:
			sendChannel <- "{\"jsonrpc\":\"2.0\",\"method\":\"Filecoin.MpoolPushMessage\",\"params\":[{\"Version\":0,\"To\":\"" + receiverAddress + "\",\"From\":\"" + senderAddress + "\",\"Value\":\"" + amount + "\",\"GasLimit\":0,\"GasFeeCap\":\"0\",\"GasPremium\":\"0\",\"Method\":0,\"CID\":{}},{\"MaxFee\":\"0\"}],\"id\":" + strconv.Itoa(sendId) + "}"
			time.Sleep(time.Duration(float64(1/rate) * 1e9))
		}

	}

}
