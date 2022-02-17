package main

import (
	"fmt"
)

/**
 * Main Benchmark invocation method
 */
func main() {

	// Check that benchmark is run as root, as commands needed to be executed as root
	if whoami() != "root" {
		fmt.Printf("NEED TO RUN AS ROOT!\n")
		return
	}

	// Get current nodes
	nodes, hosts := getCurrentNodes()

	// Define the testcases
	testcases := []TestCase{
		/*		{duration: 180, msgPerSec: 25},
				{duration: 180, msgPerSec: 50},
				{duration: 180, msgPerSec: 75},
				{duration: 180, msgPerSec: 100},
				{duration: 180, msgPerSec: 150},
				{duration: 180, msgPerSec: 200},
				{duration: 180, msgPerSec: 250},*/
		{duration: 180, msgPerSec: 300},
		{duration: 180, msgPerSec: 400},
		{duration: 180, msgPerSec: 500},
		{duration: 180, msgPerSec: 750},
		{duration: 180, msgPerSec: 1000},
		{duration: 180, msgPerSec: 1500},
		{duration: 180, msgPerSec: 2000},
		{duration: 180, msgPerSec: 2500},
	}

	// Run test
	throughputTest(nodes, hosts, testcases)

}
