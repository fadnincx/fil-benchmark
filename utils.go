package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

func Map(vs []BlockLog, f func(BlockLog) uint64) []uint64 {
	vsm := make([]uint64, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
func Max(vs []uint64) uint64 {
	m := vs[0]
	for _, v := range vs {
		if v > m {
			m = v
		}
	}
	return m
}
func Min(vs []uint64) uint64 {
	m := vs[0]
	for _, v := range vs {
		if v < m {
			m = v
		}
	}
	return m
}
func Mean(vs []uint64) uint64 {
	total := uint64(0)
	for _, v := range vs {
		total += v
	}
	return uint64(math.Round(float64(total) / float64(len(vs))))
}
func Median(vs []uint64) uint64 {
	sort.Slice(vs, func(i, j int) bool { return vs[i] < vs[j] })
	mNumber := len(vs) / 2

	if len(vs)%2 == 0 {
		return (vs[mNumber-1] + vs[mNumber]) / 2
	} else {
		return vs[mNumber]
	}
}

func websocketC(host string, interrupt chan bool, input chan string, output chan string) {

	token := remoteCmdResult(host, "lotus auth create-token --perm admin")
	// token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.IEI6vp33j8d2vTiAki55-gLGQyJZddoHfG6UT5aevPg"
	u := "ws://" + host + ":3000/rpc/v0?token=" + token
	log.Printf("connecting to %s", u)

	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func(o chan string) {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			// log.Printf("recv: %s", message)
			output <- string(message)
		}
	}(output)

	for {
		select {
		case <-done:
			return
		case t := <-input:
			err := c.WriteMessage(websocket.TextMessage, []byte(t))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			return
		default:
			time.Sleep(1)
		}
	}
}
func monitorChanOverflow(ch chan string, f func()) {
	/*for true {
		if len(ch) == cap(ch) {
			f()
		} else {
			time.Sleep(10)
		}
	}*/
}
func remoteCmdResult(ip string, cmd string) string {
	resp, err := http.Get("http://" + ip + "?" + url.PathEscape(cmd))
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return strings.TrimSuffix(string(body), "\n")
}
func cmd(command string) (error, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return err, stdout.String(), stderr.String()
}

type node struct {
	hostname string
	ip       string
	wallet   string
}

func whoami() string {
	err, out, _ := cmd("whoami")
	if err != nil {
		log.Printf("error getting pod amount: %v\n", err)
	}
	fmt.Printf("I am %v", out)
	return strings.TrimSuffix(out, "\n")
}
func getRedisNode() string {

	err, out, _ := cmd("kubectl get pods -o wide | grep lotus-redis | awk '{print $6}' | wc -l")
	if err != nil {
		log.Printf("error getting pod amount: %v\n", err)
	}
	amount, _ := strconv.Atoi(strings.TrimSuffix(out, "\n"))

	fmt.Printf("Got %v pods\n", amount)

	if amount > 0 {
		err, ip, _ := cmd(fmt.Sprintf("kubectl get pods -o wide | grep lotus-redis-0 | awk '{print $6}'"))
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		return strings.TrimSuffix(ip, "\n")
	}

	return ""
}
func getCurrentNodes() ([]node, []string) {

	err, out, _ := cmd("k3s kubectl get pods -o wide | grep lotus-node | awk '{print $6}' | wc -l")
	if err != nil {
		log.Printf("error getting pod amount: %v\n", err)
	}
	amount, _ := strconv.Atoi(strings.TrimSuffix(out, "\n"))

	fmt.Printf("Got %v pods\n", amount)

	resultNode := make([]node, amount, amount)
	resultHosts := make([]string, amount, amount)

	for i := 0; i < amount; i++ {
		err, ip, _ := cmd(fmt.Sprintf("k3s kubectl get pods -o wide | grep lotus-node-%d | awk '{print $6}'", i))
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		resultNode[i].ip = strings.TrimSuffix(ip, "\n")
		err, hostname, _ := cmd(fmt.Sprintf("k3s kubectl get pods -o wide | grep lotus-node-%d | awk '{print $1}'", i))
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		resultNode[i].hostname = strings.TrimSuffix(hostname, "\n")
		resultHosts[i] = strings.TrimSuffix(hostname, "\n")

		keyCount, _ := strconv.Atoi(remoteCmdResult(resultNode[i].ip, "lotus wallet list --addr-only | wc -l"))
		if keyCount < 1 {
			resultNode[i].wallet = remoteCmdResult(resultNode[i].ip, "lotus wallet new")
		} else {
			resultNode[i].wallet = remoteCmdResult(resultNode[i].ip, "lotus wallet list --addr-only | tail -n 1")
		}
	}

	return resultNode, resultHosts
}
