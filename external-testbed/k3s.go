package external_testbed

import (
	fil_benchmark_exec "fil-benchmark/datastructures"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type K3s struct{ ExternalTestbed }

func (_ K3s) GetLotusApiToken(node fil_benchmark_exec.Node) string {
	return remoteCmdResult(node.Ip, "lotus auth create-token --perm admin")
}
func (_ K3s) GetLotusHosts() []fil_benchmark_exec.Node {
	checkIsRoot()
	err, out, _ := LocalCmd("kubectl get pods -o=go-template='{{println `NAME IP`}}{{range .items}}{{.metadata.name}} {{.status.podIP}}{{println ``}}{{end}}' | grep lotus-node | awk '{print $2}' | wc -l")
	if err != nil {
		log.Printf("error getting pod amount: %v\n", err)
	}
	amount, _ := strconv.Atoi(strings.TrimSuffix(out, "\n"))

	fmt.Printf("Got %v pods\n", amount)

	resultNode := make([]fil_benchmark_exec.Node, amount, amount)

	for i := 0; i < amount; i++ {
		err, ip, _ := LocalCmd(fmt.Sprintf("kubectl get pods -o=go-template='{{println `NAME IP`}}{{range .items}}{{.metadata.name}} {{.status.podIP}}{{println ``}}{{end}}' | grep 'lotus-node-%d ' | awk '{print $2}'", i))
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		resultNode[i].Ip = strings.TrimSuffix(ip, "\n")
		err, hostname, _ := LocalCmd(fmt.Sprintf("kubectl get pods -o=go-template='{{println `NAME IP`}}{{range .items}}{{.metadata.name}} {{.status.podIP}}{{println ``}}{{end}}' | grep 'lotus-node-%d ' | awk '{print $1}'", i))
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		resultNode[i].Hostname = strings.TrimSuffix(hostname, "\n")

		keyCount, _ := strconv.Atoi(remoteCmdResult(resultNode[i].Ip, "lotus wallet list --addr-only | wc -l"))
		if keyCount < 1 {
			resultNode[i].SendWallet = remoteCmdResult(resultNode[i].Ip, "lotus wallet new")
		} else {
			resultNode[i].SendWallet = remoteCmdResult(resultNode[i].Ip, "lotus wallet list --addr-only | tail -n 1")
		}
	}
	return resultNode
}
func (_ K3s) GetRedisHost() string {
	checkIsRoot()

	err, out, _ := LocalCmd("kubectl get pods -o=go-template='{{println `NAME IP`}}{{range .items}}{{.metadata.name}} {{.status.podIP}}{{println ``}}{{end}}' | grep lotus-redis | awk '{print $2}' | wc -l")
	if err != nil {
		log.Printf("error getting pod amount: %v\n", err)
	}
	amount, _ := strconv.Atoi(strings.TrimSuffix(out, "\n"))

	fmt.Printf("Got %v pods\n", amount)

	if amount > 0 {
		err, ip, _ := LocalCmd(fmt.Sprintf("kubectl get pods -o=go-template='{{println `NAME IP`}}{{range .items}}{{.metadata.name}} {{.status.podIP}}{{println ``}}{{end}}' | grep lotus-redis-0 | awk '{print $2}'"))
		if err != nil {
			log.Printf("error: %v\n", err)
		}
		return strings.TrimSuffix(ip, "\n")
	}

	return ""
}

func (_ K3s) DeployTestbed(filLotusDevnetPath string, nodeCount uint64, topology string, singleBlock bool) {
	checkIsRoot()

	sb := "false"
	if singleBlock {
		sb = "true"
	}
	log.Printf("Deploy command: %s", fmt.Sprintf("cd %s && ./build_start.sh %d %s %s", filLotusDevnetPath, nodeCount, topology, sb))
	err, sout, eout := LocalCmd(fmt.Sprintf("cd %s && ./build_start.sh %d %s %s", filLotusDevnetPath, nodeCount, topology, sb))
	if err != nil {
		log.Printf("error launching testbed: %v\n", err)
	}
	fmt.Println(sout)
	fmt.Fprintln(os.Stderr, eout)
	log.Println("Deployed Testbed")
}

func (_ K3s) StopTestbed(filLotusDevnetPath string) {
	checkIsRoot()
	log.Println("Stopping Testbed")
	err, _, _ := LocalCmd(fmt.Sprintf("cd %s && ./stop.sh", filLotusDevnetPath))
	if err != nil {
		log.Printf("error stopping testbed: %v\n", err)
	}
	log.Println("Stopped Testbed")
}

func checkIsRoot() {
	if whoami() != "root" {
		fmt.Fprintf(os.Stderr, "NEED TO RUN AS ROOT!\n")
		os.Exit(2)
	}
}
func whoami() string {
	err, out, _ := LocalCmd("whoami")
	if err != nil {
		log.Printf("error getting pod amount: %v\n", err)
	}
	fmt.Printf("I am %v", out)
	return strings.TrimSuffix(out, "\n")
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
