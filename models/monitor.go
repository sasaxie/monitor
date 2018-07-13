package models

import "sync"

type Request struct {
	Addresses []string
}

type Responses struct {
	M        sync.Mutex
	Count    int64
	Response *Response
}

type Response struct {
	Data []*TableData `json:"data"`
}

type TableData struct {
	Address              string
	NowBlockNum          int64
	NowBlockHash         string
	LastSolidityBlockNum int64
	GRPC                 int64 `json:"gRPC"`
	Message              string
	TotalTransaction     int64
}

// 每新增一个socket连接，计数加一
func (r *Responses) Increase() {
	r.M.Lock()
	r.Count++
	r.M.Unlock()
}

// 每关闭一个socket连接，计数减一
func (r *Responses) Reduce() {
	r.M.Lock()
	r.Count--
	r.M.Unlock()
}

// 当计数为零时，不需要运行
func (r *Responses) Runnable() bool {
	return r.Count != 0
}
