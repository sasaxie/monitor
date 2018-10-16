package main

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"log"
	"strings"
	"sync"
	"time"
)

var c client.Client

type WitnessInfo struct {
	Info map[string]int64
	lock *sync.Mutex
}

func main() {

	var err error
	c, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: "tron",
		Password: "trondb",
	})

	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	go func() {
		ticker := time.NewTicker(40 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				for _, address := range models.Sc.Addresses {
					if strings.EqualFold("full_node", address.Type) {
						go dealFullNode(address.Ip, address.Port)
					} else if strings.EqualFold("solidity_node", address.Type) {
						go dealSolidityNode(address.Ip, address.Port)
					}
				}
			}
		}
	}()

	for {
		time.Sleep(time.Minute)
	}
}

func dealFullNode(ip string, port int) {
	address := fmt.Sprintf("%s:%d", ip, port)

	cli := service.NewFullNodeGrpcClient(address)

	cli.Start()

	defer cli.Conn.Close()

	var wg sync.WaitGroup
	var num int64 = 0
	var ping int64 = 0
	var lastSolidityBlockNum int64 = 0
	witnessInfo := new(WitnessInfo)
	witnessInfo.Info = make(map[string]int64)
	witnessInfo.lock = new(sync.Mutex)
	wg.Add(4)
	go reportFullNodeNowBlockNum(cli, &wg, &num)
	go reportFullNodePing(cli, &wg, &ping)
	go reportFullNodeLastSolidityBlockNum(cli, &wg, &lastSolidityBlockNum)
	go reportFullNodeWitness(cli, &wg, witnessInfo)
	wg.Wait()

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "tronmonitor",
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	tags := map[string]string{"node": ip}
	fields := map[string]interface{}{
		"NowBlockNum":          num,
		"ping":                 ping,
		"LastSolidityBlockNum": lastSolidityBlockNum,
	}

	pt, err := client.NewPoint("node_status", tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	tags2 := map[string]string{"node": ip}
	fields2 := make(map[string]interface{})
	witnessInfo.lock.Lock()
	for k, v := range witnessInfo.Info {
		fields2[k] = v
	}
	witnessInfo.lock.Unlock()
	pt2, err := client.NewPoint("witness", tags2, fields2, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt2)

	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}
}

func dealSolidityNode(ip string, port int) {
	address := fmt.Sprintf("%s:%d", ip, port)

	cli := service.NewSolidityNodeGrpcClient(address)

	cli.Start()

	defer cli.Conn.Close()

	var wg sync.WaitGroup
	var num int64 = 0
	var ping int64 = 0
	var lastSolidityBlockNum int64 = 0
	witnessInfo := new(WitnessInfo)
	witnessInfo.Info = make(map[string]int64)
	witnessInfo.lock = new(sync.Mutex)
	wg.Add(4)
	go reportSolidityNodeNowBlockNum(cli, &wg, &num)
	go reportSolidityNodePing(cli, &wg, &ping)
	go reportSolidityNodeLastSolidityBlockNum(cli, &wg, &lastSolidityBlockNum)
	go reportSolidityNodeWitness(cli, &wg, witnessInfo)
	wg.Wait()

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "tronmonitor",
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	tags := map[string]string{"node": ip}
	fields := map[string]interface{}{
		"NowBlockNum":          num,
		"ping":                 ping,
		"LastSolidityBlockNum": lastSolidityBlockNum,
		"TotalTransaction":     0,
	}

	pt, err := client.NewPoint("node_status", tags, fields, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	tags2 := map[string]string{"node": ip}
	fields2 := make(map[string]interface{})
	witnessInfo.lock.Lock()
	for k, v := range witnessInfo.Info {
		fields2[k] = v
	}
	witnessInfo.lock.Unlock()
	pt2, err := client.NewPoint("witness", tags2, fields2, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt2)

	if err := c.Write(bp); err != nil {
		log.Fatal(err)
	}
}

func reportFullNodeNowBlockNum(client *service.FullNodeGrpcClient,
	wg *sync.WaitGroup, num *int64) {
	defer wg.Done()
	*num = client.GetNowBlockNum()
}

func reportFullNodePing(client *service.FullNodeGrpcClient,
	wg *sync.WaitGroup, ping *int64) {
	defer wg.Done()
	*ping = client.GetPing()
}

func reportFullNodeLastSolidityBlockNum(client *service.FullNodeGrpcClient,
	wg *sync.WaitGroup, lastSolidityBlockNum *int64) {
	defer wg.Done()
	*lastSolidityBlockNum = client.GetLastSolidityBlockNum()
}

func reportFullNodeTotalTransaction(client *service.FullNodeGrpcClient,
	wg *sync.WaitGroup, totalTransaction *int64) {
	defer wg.Done()
	*totalTransaction = client.TotalTransaction()
}

func reportFullNodeWitness(client *service.FullNodeGrpcClient,
	wg *sync.WaitGroup, res *WitnessInfo) {
	defer wg.Done()
	witnessList := client.ListWitnesses()

	for _, witness := range witnessList.Witnesses {
		if witness.IsJobs {
			key := witness.Url
			res.lock.Lock()
			res.Info[key] = witness.TotalMissed
			res.lock.Unlock()
		}
	}
}

func reportSolidityNodeNowBlockNum(client *service.SolidityNodeGrpcClient,
	wg *sync.WaitGroup, num *int64) {
	defer wg.Done()
	*num = client.GetNowBlockNum()
}

func reportSolidityNodePing(client *service.SolidityNodeGrpcClient, wg *sync.WaitGroup, ping *int64) {
	defer wg.Done()
	*ping = client.GetPing()
}

func reportSolidityNodeLastSolidityBlockNum(client *service.SolidityNodeGrpcClient, wg *sync.WaitGroup, lastSolidityBlockNum *int64) {
	defer wg.Done()
	*lastSolidityBlockNum = client.GetLastSolidityBlockNum()
}

func reportSolidityNodeWitness(client *service.SolidityNodeGrpcClient,
	wg *sync.WaitGroup, res *WitnessInfo) {
	defer wg.Done()
	witnessList := client.ListWitnesses()

	for _, witness := range witnessList.Witnesses {
		if witness.IsJobs {
			key := witness.Url
			res.lock.Lock()
			res.Info[key] = witness.TotalMissed
			res.lock.Unlock()
		}
	}
}
