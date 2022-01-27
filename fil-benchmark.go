package main

import (
	"fmt"
)

func main() {

	if whoami() != "root" {
		fmt.Printf("NEED TO RUN AS ROOT!\n")
		return
	}

	nodes, hosts := getCurrentNodes()

	fmt.Println(nodes)

	throughputTest(nodes, hosts, 1, 10000, 2, 180)

	// redisGetMsgWith2Times()

}
