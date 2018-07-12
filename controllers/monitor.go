package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/service"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Operations about monitor
type MonitorController struct {
	BaseController
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
			tableData.GRPCMonitor = ""

			if pings, ok := TronMonitor.GRPCMonitor.LatestGRPCs[tableData.Address]; ok {
				for index, ping := range pings {
					tableData.GRPCMonitor += strconv.Itoa(int(ping))

					if index != len(pings)-1 {
						tableData.GRPCMonitor += ","
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

	mutex.Lock()
	client := service.GrpcClients[address]
	mutex.Unlock()

	if client != nil {
		wg.Add(1)
		go client.GetNowBlock(&tableData.NowBlockNum, &tableData.NowBlockHash, &wg)

		wg.Add(1)
		go client.GetLastSolidityBlockNum(&tableData.LastSolidityBlockNum, &wg)

		wg.Add(1)
		go GetPing(client, &tableData.GRPC, &wg)

		wg.Add(1)
		go client.TotalTransaction(&tableData.TotalTransaction, &wg)

		wg.Wait()
	}

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
	m.Data["json"] = models.Program.Runtime.Unix()

	m.ServeJSON()
}

// @Title Get server config
// @Description get server config
// @router /server-config [get]
func (m *MonitorController) ServerConfig() {
	m.Data["json"] = models.ServersConfig

	m.ServeJSON()
}

// @Title Get server config
// @Description get server config
// @router /server-config [post]
func (m *MonitorController) ServerConfigEdit() {

	server := new(models.Servers)
	err := json.Unmarshal(m.Ctx.Input.RequestBody, server)

	if err != nil {
		m.Data["json"] = err.Error()
	} else {

		if server.Servers == nil {
			m.Data["json"] = "error"
		} else {
			path := beego.AppConfig.String(models.ServerFilePath)
			file, err := os.Create(path)
			defer file.Close()

			if err != nil {
				m.Data["json"] = err.Error()
			} else {
				server.FlushToFile(file)

				oldAddresses := models.ServersConfig.GetAllAddresses()

				models.ServersConfig = server

				newAddresses := models.ServersConfig.GetAllAddresses()

				// 删掉的话，必须close
				closeOldGrpcClients(oldAddresses, newAddresses)

				// 新增的话，只重新初始化新增的
				connNewGrpcClients(oldAddresses, newAddresses)

				service.InitGrpcClients()
				m.Data["json"] = "success"
			}
		}
	}

	m.ServeJSON()
}

func closeOldGrpcClients(oldAddresses, newAddresses []string) {
	closeAddresses := make([]string, 0)

	for _, o := range oldAddresses {
		count := 0
		for _, n := range newAddresses {
			count++
			if strings.EqualFold(o, n) {
				break
			}
		}

		if count == len(newAddresses) {
			closeAddresses = append(closeAddresses, o)
		}
	}

	for _, c := range closeAddresses {
		if service.GrpcClients[c].Conn != nil {
			service.GrpcClients[c].Conn.Close()
		}
	}
}

func connNewGrpcClients(oldAddresses, newAddresses []string) {
	connAddresses := make([]string, 0)

	for _, n := range newAddresses {
		count := 0
		for _, o := range oldAddresses {
			count++
			if strings.EqualFold(o, n) {
				break
			}
		}

		if count == len(oldAddresses) {
			connAddresses = append(connAddresses, n)
		}
	}

	for _, c := range connAddresses {
		service.GrpcClients[c] = service.NewGrpcClient(c)
		service.GrpcClients[c].Start()
	}
}
