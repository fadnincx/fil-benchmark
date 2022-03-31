package external_testbed

import (
	fil_benchmark_exec "fil-benchmark/datastructures"
)

type ExternalTestbed interface {
	GetLotusApiToken(node fil_benchmark_exec.Node) string
	GetLotusHosts() []fil_benchmark_exec.Node
	GetRedisHost() string
}

func GetTestBed() ExternalTestbed {
	return K3s{}
}
