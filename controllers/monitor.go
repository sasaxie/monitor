package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"strings"
	"sync"
)

var addresses = []string{
	"18.196.99.16:50051", "18.195.254.44:50051", "18.196.78.56:50051",
	"54.236.37.243:50051", "35.169.107.157:50051", "34.237.220.206:50051", "52.53.189.99:50051", "13.57.246.69:50051", "13.57.41.129:50051", "52.15.93.92:50051", "18.217.144.24:50051", "13.58.203.73:50051", "34.220.77.106:50051", "34.220.59.202:50051", "34.216.106.30:50051", "34.253.187.192:50051", "52.16.167.215:50051", "34.250.7.238:50051", "52.56.56.149:50051", "52.56.115.243:50051", "18.130.99.124:50051", "35.180.51.163:50051", "35.180.22.225:50051", "52.47.117.230:50051", "54.252.224.209:50051", "54.252.238.51:50051", "13.211.164.189:50051", "18.228.15.36:50051", "18.231.118.237:50051", "54.232.225.66:50051", "35.182.37.246:50051", "35.183.101.7:50051", "13.229.128.108:50051", "13.229.135.228:50051", "13.124.62.58:50051", "13.125.249.129:50051", "13.127.47.162:50051", "35.154.40.248:50051",
}

// Operations about monitor
type MonitorController struct {
	beego.Controller
}

var waitGroup sync.WaitGroup
var mutex sync.Mutex

// @Title Get info
// @Description get info
// @router /info [get,post]
func (m *MonitorController) Info() {
	response := new(models.Response)
	response.Data = make([]*models.TableData, 0)

	for _, address := range addresses {
		waitGroup.Add(1)
		go getResult(address, response)
	}

	waitGroup.Wait()

	for _, tableData := range response.Data {
		if tableData.LastSolidityBlockNum == 0 {
			tableData.Message = "timeout"
		} else {
			tableData.Message = "success"
		}
	}

	m.Data["json"] = response

	m.ServeJSON()
}

func getResult(address string, response *models.Response) {
	defer waitGroup.Done()

	var wg sync.WaitGroup
	tableData := new(models.TableData)
	adds := strings.Split(address, ".")
	if len(adds) > 3 {
		tableData.Address = fmt.Sprintf("%s.*.*.%s", adds[0], adds[3])
	}

	client := service.NewGrpcClient(address)
	client.Start()
	defer client.Conn.Close()

	wg.Add(1)
	go client.GetNowBlock(&tableData.NowBlockNum, &tableData.NowBlockHash, &wg)

	wg.Add(1)
	go client.GetLastSolidityBlockNum(&tableData.LastSolidityBlockNum, &wg)

	wg.Add(1)
	go GetPing(client, &tableData.Ping, &wg)

	wg.Wait()

	mutex.Lock()
	response.Data = append(response.Data, tableData)
	mutex.Unlock()
}

func GetPing(client *service.GrpcClient, ping *int64,
	wg *sync.WaitGroup) {
	defer wg.Done()

	*ping = client.GetPing()
}
