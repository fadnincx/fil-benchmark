package datastructures

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strconv"
)

type TestCase struct {
	Name        string  `yaml:"name"`
	MsgPerSec   float64 `yaml:"msgPerSec"`
	Duration    uint64  `yaml:"duration"`
	Repetition  int     `yaml:"repetition"`
	NodeCount   uint64  `yaml:"nodeCount"`
	Topology    string  `yaml:"topology"`
	SingleBlock bool    `yaml:"singleBlock"`
}

func ReadTestcaseYaml(filename string) []TestCase {

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read testcase yml file: %s", err)
	}
	yamlFistLayerMap := make(map[string]interface{})
	err = yaml.Unmarshal(data, &yamlFistLayerMap)
	if err != nil {
		log.Fatalf("error: %test", err)
	}

	// Define Global default values
	var globalDuration uint64 = 300
	var globalRepetition int = 1
	var globalNodeCount uint64 = 8
	var globalTopology string = "star"
	var globalSingleBlock bool = false

	// Parse if default values are overwritten
	if val, ok := yamlFistLayerMap["duration"]; ok {
		duration, err := strconv.ParseUint(fmt.Sprintf("%v", val), 10, 64)
		if err == nil {
			globalDuration = duration
		}
	}
	if val, ok := yamlFistLayerMap["repetition"]; ok {
		repetition, err := strconv.Atoi(fmt.Sprintf("%v", val))
		if err == nil {
			globalRepetition = repetition
		}
	}
	if val, ok := yamlFistLayerMap["nodeCount"]; ok {
		nodeCount, err := strconv.ParseUint(fmt.Sprintf("%v", val), 10, 64)
		if err == nil {
			globalNodeCount = nodeCount
		}
	}
	if val, ok := yamlFistLayerMap["topology"]; ok {
		globalTopology = fmt.Sprintf("%test", val)
	}
	if val, ok := yamlFistLayerMap["singleBlock"]; ok {
		singleBlock, err := strconv.ParseBool(fmt.Sprintf("%v", val))
		if err == nil {
			globalSingleBlock = singleBlock
		}
	}

	// Init output slice
	testCases := make([]TestCase, 0)

	// Iterate over all tests
	for _, test := range yamlFistLayerMap["tests"].([]interface{}) {

		// Convert map[interface{}]interface{} to map[string]string for easier handling
		m := make(map[string]string)
		for k, v := range test.(map[interface{}]interface{}) {
			m[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
		}

		// init testcase struct
		var t TestCase

		//Name has to be defined, transfer to struct
		if val, ok := m["name"]; ok {
			t.Name = val
		} else {
			log.Fatalf("testcase has no name, %v", err)
		}

		// Msg Per Sec has to be defined, transfer to struct
		if val, ok := m["msgPerSec"]; ok {
			msgPerSec, _ := strconv.ParseFloat(val, 64)
			if err == nil {
				t.MsgPerSec = msgPerSec
			} else {
				log.Fatalf("testcase has no msgPerSec: %v", err)
			}
		} else {
			log.Fatalf("testcase has no msgPerSec")
		}

		// Transfer defined duration or global default to struct
		if val, ok := m["duration"]; ok {
			duration, err := strconv.ParseUint(fmt.Sprintf("%v", val), 10, 64)
			if err == nil {
				t.Duration = duration
			} else {
				t.Duration = globalDuration
			}
		} else {
			t.Duration = globalDuration
		}

		// Transfer defined repetition or global default to struct
		if val, ok := m["repetition"]; ok {
			repetition, err := strconv.Atoi(fmt.Sprintf("%v", val))
			if err == nil {
				t.Repetition = repetition
			} else {
				t.Repetition = globalRepetition
			}
		} else {
			t.Repetition = globalRepetition
		}

		// Transfer defined repetition or global default to struct
		if val, ok := m["nodeCount"]; ok {
			nodeCount, err := strconv.ParseUint(fmt.Sprintf("%v", val), 10, 64)
			if err == nil {
				t.NodeCount = nodeCount
			} else {
				t.NodeCount = globalNodeCount
			}
		} else {
			t.NodeCount = globalNodeCount
		}

		// Transfer defined topology or global default to struct
		if val, ok := m["topology"]; ok {
			t.Topology = val
		} else {
			t.Topology = globalTopology
		}

		// Transfer defined topology or global default to struct
		if val, ok := m["singleBlock"]; ok {
			singleBlock, err := strconv.ParseBool(fmt.Sprintf("%v", val))
			if err == nil {
				t.SingleBlock = singleBlock
			} else {
				t.SingleBlock = globalSingleBlock
			}
		} else {
			t.SingleBlock = globalSingleBlock
		}

		// Append test case to testcases
		testCases = append(testCases, t)

	}

	return testCases
}
