package main

import (
	"fil-benchmark/datastructures"
	testbed "fil-benchmark/external-testbed"
	benchmark "fil-benchmark/fil-benchmark-exec"
	reporting "fil-benchmark/fil-reporting"
	"fil-benchmark/utils"
	"log"
	"os"
	"path/filepath"
)

/**
 * Main Benchmark invocation method
 */
func main() {

	lotusDevnetPath := "../fil-lotus-devnet"

	// Define the testcases
	testcaseFile := "exampleTest.yml"
	if len(os.Args) == 2 {
		testcaseFile = os.Args[1]
	}
	pwd, _ := os.Getwd()
	log.Printf("Use testcase: %s\n", filepath.Join(pwd, testcaseFile))
	testcases := datastructures.ReadTestcaseYaml(filepath.Join(pwd, testcaseFile))

	for _, testcase := range testcases {

		for i := 0; i < testcase.Repetition; i++ {

			// Start testbed
			testbed.GetTestBed().DeployTestbed(lotusDevnetPath, testcase.NodeCount, testcase.Topology, testcase.SingleBlock)

			// Get current nodes
			nodes := testbed.GetTestBed().GetLotusHosts()

			// Init Reporting
			report := reporting.StartTestReport(nodes)

			// Test and init Redis
			utils.GetRedisHelper().GetClient().Get("Test").Result()

			// Run test
			benchmark.RunTestcase(testcase, nodes, report)

			// Stop testbed
			testbed.GetTestBed().StopTestbed(lotusDevnetPath)

		}

	}

}
