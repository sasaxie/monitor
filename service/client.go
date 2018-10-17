package service

import "github.com/sasaxie/monitor/api"

type Client interface {
	GetNowBlockNum() int64
	GetLastSolidityBlockNum() int64
	GetPing() int64
	ListWitnesses() *api.WitnessList
	Start()
	Shutdown()
}
