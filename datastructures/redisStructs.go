package datastructures

type TimeLogEntry struct {
	Client    string `json:"client"`
	Cid       string `json:"cid"`
	Start     int64  `json:"start"`
	ExecStart int64  `json:"exec"`
	End       int64  `json:"end"`
}
type BlockLog struct {
	Client      string   `json:"client"`
	Cid         string   `json:"cid"`
	MsgCount    uint64   `json:"amount"`
	MsgCids     []string `json:"msgcids"`
	FirstKnown  int64    `json:"firstKnown"`
	Accepted    int64    `json:"accepted"`
	TipsetEpoch string   `json:"tipsetEpoch"`
	Miner       string   `json:"miner"`
}

type BlockLogAgg struct {
	Cid         string
	MinTime     int64
	MinTimeHost int64
	FirstKnown  []int64
	Accepted    []int64
}
