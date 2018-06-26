package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/service"
	"github.com/sasaxie/monitor/models"
	"sync"
)

// Operations about monitor
type MonitorController struct {
	beego.Controller
}

var waitGroup sync.WaitGroup
var mutex sync.Mutex

// @Title Get info
// @Description get info
// @router /info [post]
func (m *MonitorController) Info() {
	response := new(models.Response)
	response.Results = make([]*models.Result, 0)

	var request models.Request
	err := json.Unmarshal(m.Ctx.Input.RequestBody, &request)

	if err != nil {
		m.Data["json"] = err.Error()
	} else {
		for _, address := range request.Addresses {
			waitGroup.Add(1)
			go getResult(address, response)
		}

		waitGroup.Wait()

		for _, v := range response.Results {
			if v.LastSolidityBlockNum == 0 {
				v.Message = "timeout"
			} else {
				v.Message = "success"
			}
		}

		m.Data["json"] = response
	}

	m.ServeJSON()
}

func getResult(address string, response *models.Response) {
	defer waitGroup.Done()

	var wg sync.WaitGroup

	result := new(models.Result)
	result.Address = address
	result.NowBlock = new(models.Block)

	client := service.NewGrpcClient(address)
	client.Start()
	defer client.Conn.Close()

	wg.Add(1)
	go client.GetNowBlock(result.NowBlock, &wg)

	wg.Add(1)
	go client.GetLastSolidityBlockNum(result, &wg)

	wg.Add(1)
	go client.GetPing(result, &wg)

	wg.Wait()

	mutex.Lock()
	response.Results = append(response.Results, result)
	mutex.Unlock()
}
