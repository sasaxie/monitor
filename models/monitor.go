package models

type Request struct {
	Addresses []string
}

type Response struct {
	Data []*TableData `json:"data"`
}

type TableData struct {
	Address              string
	NowBlockNum          int64
	NowBlockHash         string
	LastSolidityBlockNum int64
	Ping                 int64
	Message              string
}
