package collector

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/models"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

const (
	urlTemplateGetNowBlock = "http://%s:%d/%s/getnowblock"

	influxDBFieldNowBlockNode    = "Node"
	influxDBFieldNowBlockType    = "Type"
	influxDBFieldNowBlockTagName = "TagName"
	influxDBFieldNowBlockNumber  = "Number"
	influxDBPointNameNowBlock    = "api_get_now_block"
)

type GetNowBlockRequest struct {
	Common
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
	Collectors = append(Collectors, new(GetNowBlockRequest))
}

func (g *GetNowBlockRequest) Collect() {
	if !g.HasInitNodes {
		g.initNodes()
		g.HasInitNodes = true
	}

	g.start()
}

func (g *GetNowBlockRequest) initNodes() {
	if models.NodeList == nil && models.NodeList.Addresses == nil {
		panic("get now block request load() error")
	}

	if g.Nodes == nil {
		g.Nodes = make([]*Node, 0)
	}

	for _, node := range models.NodeList.Addresses {
		n := new(Node)
		n.CollectionUrl = fmt.Sprintf(
			urlTemplateGetNowBlock,
			node.Ip,
			node.HttpPort,
			config.NewNodeType(node.Type).GetApiPathByNodeType())
		n.Node = fmt.Sprintf("%s:%d", node.Ip, node.HttpPort)
		n.Type = node.Type
		n.TagName = node.TagName

		g.Nodes = append(g.Nodes, n)
	}

	logs.Info(
		"get now block request load() success, node size:",
		len(g.Nodes),
	)
}

func (g *GetNowBlockRequest) start() {
	if g.Nodes == nil || len(g.Nodes) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(g.Nodes))
	for _, node := range g.Nodes {
		go g.request(node, &wg)
	}

	wg.Wait()
}

func (g *GetNowBlockRequest) Save2db() {

}

func (g *GetNowBlockRequest) request(node *Node, wg *sync.WaitGroup) {
	defer wg.Done()

	response, err := http.Get(node.CollectionUrl)

	if err != nil {
		logs.Debug("(", node.CollectionUrl, ")", "[http get]", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		logs.Warn("get now block request (", node.CollectionUrl,
			") response status code",
			response.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[read body]", err)
		return
	}

	var block Block
	err = json.Unmarshal(body, &block)

	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[json unmarshal]", err)
		return
	}

	// Report block number
	nowBlockTags := map[string]string{
		influxDBFieldNowBlockNode:    node.Node,
		influxDBFieldNowBlockType:    node.Type,
		influxDBFieldNowBlockTagName: node.TagName,
	}

	nowBlockFields := map[string]interface{}{
		influxDBFieldNowBlockNode:    node.Node,
		influxDBFieldNowBlockType:    node.Type,
		influxDBFieldNowBlockTagName: node.TagName,

		influxDBFieldNowBlockNumber: block.BlockHeader.RawData.Number,
	}

	influxdb.Client.WriteByTime(
		influxDBPointNameNowBlock,
		nowBlockTags,
		nowBlockFields,
		time.Now(),
	)
}
