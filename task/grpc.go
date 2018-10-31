package task

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"strings"
	"sync"
	"time"
)

func StartGrpcMonitor() {
	logs.Info("start grpc monitor")
	ticker := time.NewTicker(
		time.Duration(config.MonitorConfig.Task.GetGRPCDataInterval) *
			time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, address := range models.NodeList.Addresses {
				if strings.EqualFold(config.FullNode.String(), address.Type) {
					go dealGrpcMonitor(address.Type, address.Ip, address.GrpcPort)
				} else if strings.EqualFold(config.SolidityNode.String(),
					address.Type) {
					go dealGrpcMonitor(address.Type, address.Ip, address.GrpcPort)
				}
			}
		}
	}
}

func dealGrpcMonitor(t, ip string, port int) {
	address := fmt.Sprintf("%s:%d", ip, port)

	cli := newGrpcClient(t, address)
	cli.Start()
	defer cli.Shutdown()

	var wg sync.WaitGroup
	var num int64 = 0
	var ping int64 = 0
	var lastSolidityBlockNum int64 = 0
	witnessInfo := new(models.WitnessInfo)
	witnessInfo.Info = make(map[string]int64)
	witnessInfo.Lock = new(sync.Mutex)
	wg.Add(4)
	go getNowBlockNum(cli, &wg, &num)
	go getPing(cli, &wg, &ping)
	go getLastSolidityBlockNum(cli, &wg, &lastSolidityBlockNum)
	go getWitnessList(cli, &wg, witnessInfo)
	wg.Wait()

	nodeStatusTags := map[string]string{config.InfluxDBTagNode: address}
	nodeStatusFields := map[string]interface{}{
		config.InfluxDBFieldNowBlockNum:          num,
		config.InfluxDBFieldPing:                 ping,
		config.InfluxDBFieldLastSolidityBlockNum: lastSolidityBlockNum,
	}

	witnessTags := map[string]string{config.InfluxDBTagNode: address}
	witnessFields := make(map[string]interface{})
	witnessInfo.Lock.Lock()
	for k, v := range witnessInfo.Info {
		witnessFields[k] = v
	}
	witnessInfo.Lock.Unlock()

	influxdb.Client.Write(config.InfluxDBPointNameNodeStatus, nodeStatusTags,
		nodeStatusFields)

	if len(witnessFields) > 0 {
		influxdb.Client.Write(config.InfluxDBPointNameWitness, witnessTags, witnessFields)
	}
}

func newGrpcClient(t, addr string) service.Client {
	if strings.EqualFold(t, config.FullNode.String()) {
		return service.NewFullNodeGrpcClient(addr)
	} else if strings.EqualFold(t, config.SolidityNode.String()) {
		return service.NewSolidityNodeGrpcClient(addr)
	}

	return nil
}

func getNowBlockNum(c service.Client,
	wg *sync.WaitGroup, num *int64) {
	defer wg.Done()
	*num = c.GetNowBlockNum()
}

func getPing(c service.Client,
	wg *sync.WaitGroup, ping *int64) {
	defer wg.Done()
	*ping = c.GetPing()
}

func getLastSolidityBlockNum(c service.Client,
	wg *sync.WaitGroup, lastSolidityBlockNum *int64) {
	defer wg.Done()
	*lastSolidityBlockNum = c.GetLastSolidityBlockNum()
}

func getWitnessList(c service.Client,
	wg *sync.WaitGroup, res *models.WitnessInfo) {
	defer wg.Done()
	witnessList := c.ListWitnesses()

	if witnessList != nil {
		for _, witness := range witnessList.Witnesses {
			if witness.IsJobs {
				key := witness.Url
				res.Lock.Lock()
				res.Info[key] = witness.TotalMissed
				res.Lock.Unlock()
			}
		}
	}

}
