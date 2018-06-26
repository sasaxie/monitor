package models

type Request struct {
	Addresses []string
}

type Response struct {
	Results []*Result
}

type Result struct {
	Address              string
	NowBlock             *Block
	LastSolidityBlockNum int64
	Ping                 int64
	Message              string
}

type Block struct {
	Hash   string
	Number int64
}
