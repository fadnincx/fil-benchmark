package datastructures

type TimeLogEntry struct {
	Client    string `json:"client"`
	Cid       string `json:"cid"`
	Start     int64  `json:"start"`
	ExecStart int64  `json:"exec"`
	End       int64  `json:"end"`
}
type BlockLog struct {
	Client   string `json:"client"`
	Cid      string `json:"cid"`
	MsgCount uint64 `json:"Amount"`
	Time     int64  `json:"time"`
}
