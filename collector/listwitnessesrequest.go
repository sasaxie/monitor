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

type ListWitnessesRequest struct {
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
	Collectors = append(Collectors, new(ListWitnessesRequest))
}

func (g *ListWitnessesRequest) Collect() {
	if !g.HasInitNodes {
		g.initNodes()
		g.HasInitNodes = true
	}

	g.start()
}

func (l *ListWitnessesRequest) initNodes() {
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

func (l *ListWitnessesRequest) start() {
	if l.Nodes == nil || len(l.Nodes) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(l.Nodes))
	for _, node := range l.Nodes {
		go l.request(node, &wg)
	}

	wg.Wait()
}

func (l *ListWitnessesRequest) Save2db() {

}

func (l *ListWitnessesRequest) request(node *Node, wg *sync.WaitGroup) {
	defer wg.Done()

	response, err := http.Get(node.CollectionUrl)

	if err != nil {
		logs.Debug("(", node.CollectionUrl, ")", "[http get]", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		logs.Warn("list witnesses request (", node.CollectionUrl,
			") response status code",
			response.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[read body]", err)
		return
	}

	var witnesses WitnessList
	err = json.Unmarshal(body, &witnesses)

	if err != nil {
		logs.Warn("(", node.CollectionUrl, ") ", "[json unmarshal]", err)
		return
	}

	// Report witness
	t := time.Now()
	if witnesses.Witnesses != nil {
		for _, w := range witnesses.Witnesses {
			if w.IsJobs {
				witnessTags := map[string]string{
					influxDBFieldListWitnessesNode:    node.Node,
					influxDBFieldListWitnessesType:    node.Type,
					influxDBFieldListWitnessesTagName: node.TagName,
					influxDBFieldListWitnessesUrl:     w.Url,
				}

				witnessFields := map[string]interface{}{
					influxDBFieldListWitnessesNode:    node.Node,
					influxDBFieldListWitnessesType:    node.Type,
					influxDBFieldListWitnessesTagName: node.TagName,

					influxDBFieldListWitnessesAddress:     w.Address,
					influxDBFieldListWitnessesTotalMissed: w.TotalMissed,
					influxDBFieldListWitnessesUrl:         w.Url,
					influxDBFieldListWitnessesIsJobs:      w.IsJobs,
				}

				influxdb.Client.WriteByTime(
					influxDBPointNameListWitnesses,
					witnessTags,
					witnessFields,
					t,
				)
			}
		}
	}
}
