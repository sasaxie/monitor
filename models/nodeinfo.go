package models

type NodeInfo struct {
	Address              string
	NowBlockNum          int64
	NowBlockHash         string
	LastSolidityBlockNum int64
	RequestTimeConsuming int64
	Message              string
	TotalTransaction     int64
}
