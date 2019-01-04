package collector

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/models"
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

type GetNowBlockCollector struct {
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
	Collectors = append(Collectors, new(GetNowBlockCollector))
}

func (g *GetNowBlockCollector) Collect() {
	g.init()

	g.collect()
}

func (g *GetNowBlockCollector) init() {
	if !g.HasInit {
		g.initNodes()
		g.HasInit = true
		logs.Info("init GetNowBlockCollector")
	}
}

func (g *GetNowBlockCollector) initNodes() {
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

func (g *GetNowBlockCollector) collect() {
	if g.Nodes == nil || len(g.Nodes) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(g.Nodes))
	for _, node := range g.Nodes {
		go g.collectByNode(node, &wg)
	}

	wg.Wait()
}

func (g *GetNowBlockCollector) collectByNode(node *Node, wg *sync.WaitGroup) {
	defer wg.Done()

	data, err := fetch(node.CollectionUrl)
	if err != nil {
		logs.Debug(err)
		return
	}

	var block Block
	err = json.Unmarshal(data, &block)

	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[json unmarshal]", err)
		return
	}

	g.saveBlock(node.Node, node.Type, node.TagName, block.BlockHeader.RawData.Number)

}

func (g *GetNowBlockCollector) saveBlock(nodeHost, nodeType,
	nodeTagName string, blockNum int64) {

	nowBlockTags := map[string]string{
		influxDBFieldNowBlockNode:    nodeHost,
		influxDBFieldNowBlockType:    nodeType,
		influxDBFieldNowBlockTagName: nodeTagName,
	}

	nowBlockFields := map[string]interface{}{
		influxDBFieldNowBlockNode:    nodeHost,
		influxDBFieldNowBlockType:    nodeType,
		influxDBFieldNowBlockTagName: nodeTagName,

		influxDBFieldNowBlockNumber: blockNum,
	}

	err := influxdb.Client.WriteByTime(
		influxDBPointNameNowBlock,
		nowBlockTags,
		nowBlockFields,
		time.Now(),
	)

	if err != nil {
		logs.Error(err)
	}
}
