package service

import (
	"context"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/api"
	"google.golang.org/grpc"
	"time"
)

const GrpcTimeout = 30 * time.Second

type FullNodeGrpcClient struct {
	Address        string
	Conn           *grpc.ClientConn
	WalletClient   api.WalletClient
	DatabaseClient api.DatabaseClient
}

func NewFullNodeGrpcClient(address string) *FullNodeGrpcClient {
	client := new(FullNodeGrpcClient)
	client.Address = address
	return client
}

func (g *FullNodeGrpcClient) Start() {
	var err error
	g.Conn, err = grpc.Dial(g.Address, grpc.WithInsecure())
	if err != nil {
		logs.Warn("did not connect: %v\n", err)
		return
	}

	g.WalletClient = api.NewWalletClient(g.Conn)
	g.DatabaseClient = api.NewDatabaseClient(g.Conn)
}

func (g *FullNodeGrpcClient) Shutdown() {
	g.Conn.Close()
}

func (g *FullNodeGrpcClient) GetNowBlockNum() int64 {
	if g.Conn == nil {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()
	result, err := g.WalletClient.GetNowBlock2(ctx, new(api.EmptyMessage))

	if err != nil {
		logs.Warn("get now block error: %v\n", err)
		return 0
	}

	return result.GetBlockHeader().GetRawData().GetNumber()
}

func (g *FullNodeGrpcClient) GetLastSolidityBlockNum() int64 {
	if g.Conn == nil {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()

	dynamicProperties, err := g.DatabaseClient.GetDynamicProperties(ctx, new(api.EmptyMessage))

	if err != nil {
		logs.Warn("get last solidity block num error: %v", err)
		return 0
	}

	return dynamicProperties.LastSolidityBlockNum
}

func (g *FullNodeGrpcClient) GetPing() int64 {
	if g.Conn == nil {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()

	start := time.Now().UnixNano() / 1000000
	_, err := g.WalletClient.GetNowBlock2(ctx, new(api.EmptyMessage))
	end := time.Now().UnixNano() / 1000000

	if err != nil {
		logs.Warn("get ping error: %v", err)
		return 0
	}

	return end - start
}

func (g *FullNodeGrpcClient) ListWitnesses() *api.WitnessList {
	if g.Conn == nil {
		return new(api.WitnessList)
	}

	witnessList, err := g.WalletClient.ListWitnesses(context.Background(),
		new(api.EmptyMessage))

	if err != nil {
		logs.Warn("get witnesses error: %v\n", err)
	}

	return witnessList
}
