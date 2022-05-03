
# FIL-Benchmark

This Repository is part of the master thesis of Marcel Würsten \"Filecoin Consensus Performance Analysis\"


## Usage

### Requirements
You need a running Filecoin Testnet reporting statistics to Redis [See this Repo](https://github.com/fadnincx/fil-lotus-devnet)

If you want/need to build the test, [Golang 1.18](https://go.dev/) is required

### Testcases
Define your testcase in a yml following the [`example.yml`](example.yml)

### Run the testcase
Once the test network is running in the desired configuration, the tests can be started as follows:
```bash
./fil-benchmark example.yml
```
or if you need to compile the test suite
```bash
go run . example.yml
```