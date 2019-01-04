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
	urlTemplateListWitnesses = "http://%s:%d/%s/listwitnesses"

	influxDBFieldListWitnessesNode        = "Node"
	influxDBFieldListWitnessesType        = "Type"
	influxDBFieldListWitnessesTagName     = "TagName"
	influxDBFieldListWitnessesAddress     = "Address"
	influxDBFieldListWitnessesTotalMissed = "TotalMissed"
	influxDBFieldListWitnessesUrl         = "Url"
	influxDBFieldListWitnessesIsJobs      = "IsJobs"
	influxDBPointNameListWitnesses        = "api_list_witnesses"
)

type ListWitnessesCollector struct {
	Common
}

type WitnessList struct {
	Witnesses []*Witness `json:"witnesses"`
}

type Witness struct {
	Address     string `json:"address"`
	Url         string `json:"url"`
	TotalMissed int64  `json:"totalMissed"`
	IsJobs      bool   `json:"isJobs"`
}

func init() {
	Collectors = append(Collectors, new(ListWitnessesCollector))
}

func (l *ListWitnessesCollector) Collect() {
	l.init()

	l.collect()
}

func (l *ListWitnessesCollector) init() {
	if !l.HasInit {
		l.initNodes()
		l.HasInit = true
		logs.Info("init ListWitnessesCollector")
	}
}

func (l *ListWitnessesCollector) initNodes() {
	if models.NodeList == nil && models.NodeList.Addresses == nil {
		panic("list witnesses request load() error")
	}

	if l.Nodes == nil {
		l.Nodes = make([]*Node, 0)
	}

	for _, node := range models.NodeList.Addresses {
		n := new(Node)
		n.CollectionUrl = fmt.Sprintf(
			urlTemplateListWitnesses,
			node.Ip,
			node.HttpPort,
			config.NewNodeType(node.Type).GetApiPathByNodeType())
		n.Node = fmt.Sprintf("%s:%d", node.Ip, node.HttpPort)
		n.Type = node.Type
		n.TagName = node.TagName

		l.Nodes = append(l.Nodes, n)
	}

	logs.Info(
		"list witnesses request load() success, node size:",
		len(l.Nodes),
	)
}

func (l *ListWitnessesCollector) collect() {
	if l.Nodes == nil || len(l.Nodes) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(l.Nodes))
	for _, node := range l.Nodes {
		go l.collectByNode(node, &wg)
	}

	wg.Wait()
}

func (l *ListWitnessesCollector) collectByNode(node *Node, wg *sync.WaitGroup) {
	defer wg.Done()

	data, err := fetch(node.CollectionUrl)
	if err != nil {
		logs.Debug(err)
		return
	}

	var witnesses WitnessList
	err = json.Unmarshal(data, &witnesses)

	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[json unmarshal]", err)
		return
	}

	l.saveWitness(witnesses, node.Node, node.Type, node.TagName)
}

func (l *ListWitnessesCollector) saveWitness(
	witnesses WitnessList,
	nodeHost, nodeType, nodeTagName string) {
	if witnesses.Witnesses != nil {
		for _, w := range witnesses.Witnesses {
			if w.IsJobs {
				witnessTags := map[string]string{
					influxDBFieldListWitnessesNode:    nodeHost,
					influxDBFieldListWitnessesType:    nodeType,
					influxDBFieldListWitnessesTagName: nodeTagName,
					influxDBFieldListWitnessesUrl:     w.Url,
				}

				witnessFields := map[string]interface{}{
					influxDBFieldListWitnessesNode:    nodeHost,
					influxDBFieldListWitnessesType:    nodeType,
					influxDBFieldListWitnessesTagName: nodeTagName,

					influxDBFieldListWitnessesAddress:     w.Address,
					influxDBFieldListWitnessesTotalMissed: w.TotalMissed,
					influxDBFieldListWitnessesUrl:         w.Url,
					influxDBFieldListWitnessesIsJobs:      w.IsJobs,
				}

				influxdb.Client.WriteByTime(
					influxDBPointNameListWitnesses,
					witnessTags,
					witnessFields,
					time.Now(),
				)
			}
		}
	}
}
