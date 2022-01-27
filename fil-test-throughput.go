package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func throughputTest(nodes []node, hosts []string, startMsgPerSec float64, stopMsgPerSec float64, increaseStep float64, secondsPerStep uint64) {

	inputChan := make(chan string)
	outputChan := make(chan string)
	interuptChan := make(chan bool)

	cidSend := make(chan string)
	cidAgg := make(chan []string)
	cidDoAgg := make(chan bool)

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
			cidSend <- cid

		}
	}(outputChan, cidSend)
	go websocketC(nodes[0].ip, interuptChan, inputChan, outputChan)

	currentRate := startMsgPerSec
	for currentRate < stopMsgPerSec {
		stopRateChan := make(chan bool)

		go sendAtRate(inputChan, stopRateChan, currentRate, nodes[0].wallet, nodes[1].wallet, "10")
		time.Sleep(time.Duration(secondsPerStep * 1e9))
		fmt.Println("Increase Msg per Sec")
		stopRateChan <- true

		cidDoAgg <- true
		cids := <-cidAgg

		go func(cids []string, hosts []string, rate float64) {
			amount, avg, unfinished := redisGetAvgDelay(cids, hosts)
			writeCsvLine(fmt.Sprintf("%f, %d, %f, %d\n", currentRate, amount, avg, unfinished))
			fmt.Printf("Rate %v has %v messages with %vus delay\n", currentRate, amount, avg)

		}(cids, hosts, currentRate)

		currentRate += increaseStep
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
