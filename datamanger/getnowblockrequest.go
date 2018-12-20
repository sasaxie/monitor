package datamanger

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/models"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	urlTemplateGetNowBlock = "http://%s:%d/wallet/getnowblock"

	influxDBFieldNowBlockNode   = "Node"
	influxDBFieldNowBlockType   = "Type"
	influxDBFieldNowBlockTag    = "Tag"
	influxDBFieldNowBlockNumber = "Number"
	influxDBPointNameNowBlock   = "api_get_now_block"
)

type GetNowBlockRequest struct {
	RequestCommon
}

type Block struct {
	BlockHeader *BlockHeader `json:"block_header"`
}

type BlockHeader struct {
	RawData *RawData `json:"raw_data"`
}

type RawData struct {
	Number int64 `json:"number"`
}

func init() {
	Requests = append(Requests, new(GetNowBlockRequest))
}

func (g *GetNowBlockRequest) Load() {
	if models.NodeList == nil && models.NodeList.Addresses == nil {
		panic("Get now block request load() error")
	}

	if g.Parameters == nil {
		g.Parameters = make([]*Parameter, 0)
	}

	for _, node := range models.NodeList.Addresses {
		param := new(Parameter)
		param.RequestUrl = fmt.Sprintf(urlTemplateGetNowBlock, node.Ip, node.HttpPort)
		param.Node = fmt.Sprintf("%s:%d", node.Ip, node.HttpPort)
		param.Type = node.Type
		param.Tag = node.Tag

		g.Parameters = append(g.Parameters, param)
	}

	logs.Info(
		"Get now block request load() success, node size:",
		len(g.Parameters),
	)
}

func (g *GetNowBlockRequest) Request() {
	if g.Parameters == nil || len(g.Parameters) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(g.Parameters))
	for _, param := range g.Parameters {
		go g.request(param, &wg)
	}

	wg.Wait()
}

func (g *GetNowBlockRequest) Save2db() {

}

func (g *GetNowBlockRequest) request(param *Parameter, wg *sync.WaitGroup) {
	defer wg.Done()

	response, err := http.Get(param.RequestUrl)

	if err != nil {
		logs.Debug("(", param.RequestUrl, ")", "[http get]", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		logs.Warn("Get now block request (", param.RequestUrl, ") response status code",
			response.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Warn("(", param.RequestUrl, ") ", "[read body]", err)
		return
	}

	var block Block
	err = json.Unmarshal(body, &block)

	if err != nil {
		logs.Warn("(", param.RequestUrl, ") ", "[json unmarshal]", err)
		return
	}

	// Report block number
	nowBlockTags := map[string]string{
		influxDBFieldNowBlockNode: param.Node,
		influxDBFieldNowBlockType: param.Type,
		influxDBFieldNowBlockTag:  param.Tag,
	}

	nowBlockFields := map[string]interface{}{
		influxDBFieldNowBlockNode: param.Node,
		influxDBFieldNowBlockType: param.Type,
		influxDBFieldNowBlockTag:  param.Tag,

		influxDBFieldNowBlockNumber: block.BlockHeader.RawData.Number,
	}

	influxdb.Client.WriteByTime(
		influxDBPointNameNowBlock,
		nowBlockTags,
		nowBlockFields,
		time.Now(),
	)
}
