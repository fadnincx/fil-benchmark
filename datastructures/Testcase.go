package datastructures

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type TestCase struct {
	Name      string  `yaml:"name"`
	MsgPerSec float64 `yaml:"msgPerSecPer"`
	Duration  uint64  `yaml:"duration"`
}

func ReadTestcaseYaml(filename string) []TestCase {

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read testcase yml file: %s", err)
	}
	y := make(map[string][]TestCase)
	err = yaml.Unmarshal(data, &y)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	caseAmount := len(y["tests"])
	if caseAmount < 1 {
		log.Fatalf("No test cases defined!")
	}

	return y["tests"]
}
