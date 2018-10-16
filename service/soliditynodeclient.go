package service

import (
	"context"
	"github.com/sasaxie/monitor/api"
	"google.golang.org/grpc"
	"log"
	"time"
)

type SolidityNodeGrpcClient struct {
	Address        string
	Conn           *grpc.ClientConn
	WalletClient   api.WalletSolidityClient
	DatabaseClient api.DatabaseClient
}

func NewSolidityNodeGrpcClient(address string) *SolidityNodeGrpcClient {
	client := new(SolidityNodeGrpcClient)
	client.Address = address
	return client
}

func (g *SolidityNodeGrpcClient) Start() {
	var err error
	g.Conn, err = grpc.Dial(g.Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}

	g.WalletClient = api.NewWalletSolidityClient(g.Conn)
	g.DatabaseClient = api.NewDatabaseClient(g.Conn)
}

func (g *SolidityNodeGrpcClient) GetNowBlockNum() int64 {
	if g.Conn == nil {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()
	result, err := g.WalletClient.GetNowBlock2(ctx, new(api.EmptyMessage))

	if err != nil {
		log.Printf("get now block error: %v\n", err)
		return 0
	}

	return result.GetBlockHeader().GetRawData().GetNumber()
}

func (g *SolidityNodeGrpcClient) GetLastSolidityBlockNum() int64 {
	if g.Conn == nil {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()

	dynamicProperties, err := g.DatabaseClient.GetDynamicProperties(ctx, new(api.EmptyMessage))

	if err != nil {
		log.Printf("get last solidity block num error: %v", err)
		return 0
	}

	return dynamicProperties.LastSolidityBlockNum
}

func (g *SolidityNodeGrpcClient) GetPing() int64 {
	if g.Conn == nil {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()

	start := time.Now().UnixNano() / 1000000
	_, err := g.WalletClient.GetNowBlock2(ctx, new(api.EmptyMessage))
	end := time.Now().UnixNano() / 1000000

	if err != nil {
		log.Printf("get ping error: %v", err)
		return 0
	}

	return end - start
}

func (g *SolidityNodeGrpcClient) ListWitnesses() *api.WitnessList {
	if g.Conn == nil {
		return new(api.WitnessList)
	}

	witnessList, err := g.WalletClient.ListWitnesses(context.Background(),
		new(api.EmptyMessage))

	if err != nil {
		log.Printf("get witnesses error: %v\n", err)
	}

	return witnessList
}
