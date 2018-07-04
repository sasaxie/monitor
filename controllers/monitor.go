package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"strconv"
	"sync"
	"time"
)

// Operations about monitor
type MonitorController struct {
	beego.Controller
}

var waitGroup sync.WaitGroup
var mutex sync.Mutex

// @Title Get info
// @Description get info
// @router /info/tag/:tag [get,post]
func (m *MonitorController) Info() {
	tag := m.GetString(":tag")

	if tag == "" && len(tag) == 0 {
		m.Data["json"] = "not found tag"
	} else {
		addresses := models.ServersConfig.GetAddressStringByTag(tag)

		response := new(models.Response)
		response.Data = make([]*models.TableData, 0)

		for _, address := range addresses {
			waitGroup.Add(1)
			go getResult(address, response)
		}

		waitGroup.Wait()

		for _, tableData := range response.Data {
			tableData.PingMonitor = ""

			if pings, ok := PingMonitor[tableData.Address]; ok {
				for index, ping := range pings {
					tableData.PingMonitor += strconv.Itoa(int(ping))

					if index != len(pings)-1 {
						tableData.PingMonitor += ","
					}
				}
			}

			if tableData.LastSolidityBlockNum == 0 {
				tableData.Message = "timeout"
			} else {
				tableData.Message = "success"
			}
		}

		m.Data["json"] = response
	}

	m.ServeJSON()
}

func getResult(address string, response *models.Response) {
	defer waitGroup.Done()

	var wg sync.WaitGroup
	tableData := new(models.TableData)
	tableData.Address = address

	client := service.GrpcClients[address]

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

// @Title Get tags
// @Description get tags
// @router /tags [get,post]
func (m *MonitorController) Tags() {
	m.Data["json"] = models.ServersConfig.GetTags()
	m.ServeJSON()
}

// @Title Get settings
// @Description get settings
// @router /settings [get,post]
func (m *MonitorController) Settings() {
	m.Data["json"] = models.ServersConfig.GetSettings()
	m.ServeJSON()
}

// @Title Get program info
// @Description get program info
// @router /program-info [get,post]
func (m *MonitorController) ProgramInfo() {
	now := time.Now().UTC().Unix()

	duration := now - models.Program.Runtime.Unix()

	d, err := time.ParseDuration(fmt.Sprintf("%ds", duration))

	if err != nil {
		m.Data["json"] = err.Error()
	} else {
		m.Data["json"] = d.String()
	}

	m.ServeJSON()
}
