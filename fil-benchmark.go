package main

import (
	"fil-benchmark/datastructures"
	testbed "fil-benchmark/external-testbed"
	benchmark "fil-benchmark/fil-benchmark-exec"
	reporting "fil-benchmark/fil-reporting"
	"log"
	"os"
	"path/filepath"
)

/**
 * Main Benchmark invocation method
 */
func main() {

	// Get current nodes
	nodes := testbed.GetTestBed().GetLotusHosts()

	// Define the testcases

	testcaseFile := "exampleTest.yml"
	if len(os.Args) == 2 {
		testcaseFile = os.Args[1]
	}
	pwd, _ := os.Getwd()
	log.Printf("Use testcase: %s\n", filepath.Join(pwd, testcaseFile))
	testcases := datastructures.ReadTestcaseYaml(filepath.Join(pwd, testcaseFile))

	// Init Reporting
	report := reporting.StartTestReport(nodes)

	// Run test
	benchmark.RunTestcases(testcases, nodes, report)

}
