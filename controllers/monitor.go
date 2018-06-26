package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/common/hexutil"
	"github.com/sasaxie/monitor/service"
	"github.com/sasaxie/monitor/util"
	"strings"
	"time"
)

// Operations about monitor
type MonitorController struct {
	beego.Controller
}

type Request struct {
	Addresses []string
}

type Response struct {
	Results []Result
}

type Result struct {
	Address              string
	NowBlock             Block
	LastSolidityBlockNum int64
	Ping                 int64
	Message              string
}

type Block struct {
	Hash   string
	Number int64
}

// @Title Get info
// @Description get info
// @router /info [post]
func (m *MonitorController) Info() {
	response := new(Response)
	response.Results = make([]Result, 0)

	var request Request
	err := json.Unmarshal(m.Ctx.Input.RequestBody, &request)

	if err != nil {
		m.Data["json"] = err.Error()
	} else {
		for _, address := range request.Addresses {
			go getResult(address, response)
		}
		m.Data["json"] = response
	}

	time.Sleep(1000 * time.Millisecond)

	for _, address := range request.Addresses {
		for i, a := range response.Results {
			if strings.EqualFold(a.Address, address) {
				response.Results[i].Message = "success"
				break
			}
		}
	}

	for _, address := range request.Addresses {
		isHave := false
		for _, a := range response.Results {
			if strings.EqualFold(a.Address, address) {
				isHave = true
				break
			}
		}

		if !isHave {
			var res Result
			res.Message = "timeout"
			res.Address = address
			response.Results = append(response.Results, res)
		}
	}

	m.ServeJSON()
}

func getResult(address string, response *Response) {
	var result Result
	result.Address = address
	var nowBlock Block

	client := service.NewGrpcClient(address)
	client.Start()
	defer client.Conn.Close()

	start := time.Now().UnixNano() / 1000000
	block := client.GetNowBlock()
	result.Ping = time.Now().UnixNano()/1000000 - start

	if block != nil {
		nowBlock.Hash = hexutil.Encode(util.GetBlockHash(*block))
		nowBlock.Number = block.GetBlockHeader().GetRawData().GetNumber()
		result.NowBlock = nowBlock

		result.LastSolidityBlockNum = client.GetLastSolidityBlockNum()

		response.Results = append(response.Results, result)
	}
}
