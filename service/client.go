package service

import (
	"context"
	"github.com/sasaxie/monitor/api"
	"github.com/sasaxie/monitor/common/hexutil"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/util"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
)

const GrpcTimeout = 5 * time.Second

var GrpcClients map[string]*GrpcClient

func InitGrpcClients() {
	GrpcClients = make(map[string]*GrpcClient)

	addresses := models.ServersConfig.GetAllAddresses()

	for _, address := range addresses {
		grcpClient := NewGrpcClient(address)
		grcpClient.Start()
		GrpcClients[address] = grcpClient
	}
}

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

func (g *GrpcClient) GetNowBlock(num *int64, hash *string, wg *sync.WaitGroup) {
	defer wg.Done()

	if g.Conn == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()
	result, err := g.WalletClient.GetNowBlock(ctx, new(api.EmptyMessage))

	if err != nil {
		log.Printf("get now block error: %v\n", err)
		return
	}

	*hash = hexutil.Encode(util.GetBlockHash(*result))
	*num = result.GetBlockHeader().GetRawData().GetNumber()
}

func (g *GrpcClient) GetNextMaintenanceTime() *api.NumberMessage {
	if g.Conn == nil {
		return new(api.NumberMessage)
	}

	result, err := g.WalletClient.GetNextMaintenanceTime(context.Background(),
		new(api.EmptyMessage))

	if err != nil {
		log.Fatalf("get next maintenance time error: %v", err)
	}

	return result
}

func (g *GrpcClient) TotalTransaction() *api.NumberMessage {
	if g.Conn == nil {
		return new(api.NumberMessage)
	}

	result, err := g.WalletClient.TotalTransaction(context.Background(),
		new(api.EmptyMessage))

	if err != nil {
		log.Fatalf("total transaction error: %v", err)
	}

	return result
}

func (g *GrpcClient) GetLastSolidityBlockNum(num *int64,
	wg *sync.WaitGroup) {
	defer wg.Done()

	if g.Conn == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()

	dynamicProperties, err := g.DatabaseClient.GetDynamicProperties(ctx, new(api.EmptyMessage))

	if err != nil {
		log.Printf("get last solidity block num error: %v", err)
		return
	}

	*num = dynamicProperties.LastSolidityBlockNum
}

func (g *GrpcClient) GetPing() int64 {
	if g.Conn == nil {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), GrpcTimeout)
	defer cancel()

	start := time.Now().UnixNano() / 1000000
	_, err := g.WalletClient.GetNowBlock(ctx, new(api.EmptyMessage))
	end := time.Now().UnixNano() / 1000000

	if err != nil {
		log.Printf("get ping error: %v", err)
		return 0
	}

	return end - start
}

func (g *GrpcClient) ListWitnesses() *api.WitnessList {
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
