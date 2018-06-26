package service

import (
	"context"
	"github.com/sasaxie/monitor/api"
	"github.com/sasaxie/monitor/common/hexutil"
	"google.golang.org/grpc"
	"log"
	"github.com/sasaxie/monitor/util"
	"github.com/sasaxie/monitor/models"
	"sync"
	"time"
)

const GrpcTimeout = 5 * time.Second

type GrpcClient struct {
	Address        string
	Conn           *grpc.ClientConn
	WalletClient   api.WalletClient
	DatabaseClient api.DatabaseClient
}

func NewGrpcClient(address string) *GrpcClient {
	client := new(GrpcClient)
	client.Address = address
	return client
}

func (g *GrpcClient) Start() {
	var err error
	g.Conn, err = grpc.Dial(g.Address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v\n", err)
	}

	g.WalletClient = api.NewWalletClient(g.Conn)
	g.DatabaseClient = api.NewDatabaseClient(g.Conn)
}

func (g *GrpcClient) GetNowBlock(block *models.Block, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()
	result, err := g.WalletClient.GetNowBlock(ctx, new(api.EmptyMessage))

	if err != nil {
		log.Printf("get now block error: %v\n", err)
		return
	}

	block.Hash = hexutil.Encode(util.GetBlockHash(*result))
	block.Number = result.GetBlockHeader().GetRawData().GetNumber()
}

func (g *GrpcClient) GetNextMaintenanceTime() *api.NumberMessage {

	result, err := g.WalletClient.GetNextMaintenanceTime(context.Background(),
		new(api.EmptyMessage))

	if err != nil {
		log.Fatalf("get next maintenance time error: %v", err)
	}

	return result
}

func (g *GrpcClient) TotalTransaction() *api.NumberMessage {

	result, err := g.WalletClient.TotalTransaction(context.Background(),
		new(api.EmptyMessage))

	if err != nil {
		log.Fatalf("total transaction error: %v", err)
	}

	return result
}

func (g *GrpcClient) GetLastSolidityBlockNum(result *models.Result,
	wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()

	dynamicProperties, err := g.DatabaseClient.GetDynamicProperties(ctx, new(api.EmptyMessage))

	if err != nil {
		log.Printf("get last solidity block num error: %v", err)
		return
	}

	result.LastSolidityBlockNum = dynamicProperties.LastSolidityBlockNum
}

func (g *GrpcClient) GetPing(result *models.Result, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()

	start := time.Now().UnixNano() / 1000000
	_, err := g.WalletClient.GetNowBlock(ctx, new(api.EmptyMessage))
	end := time.Now().UnixNano() / 1000000

	if err != nil {
		log.Printf("get ping error: %v", err)
		return
	}

	result.Ping = end - start
}