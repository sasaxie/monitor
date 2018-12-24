package datamanger

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
	influxDBFieldListWitnessesTag         = "Tag"
	influxDBFieldListWitnessesTotalMissed = "TotalMissed"
	influxDBFieldListWitnessesUrl         = "Url"
	influxDBPointNameListWitnesses        = "api_list_witnesses"
)

type ListWitnessesRequest struct {
	RequestCommon
}

type WitnessList struct {
	Witnesses []*Witness `json:"witnesses"`
}

type Witness struct {
	Url         string `json:"url"`
	TotalMissed int64  `json:"totalMissed"`
}

func init() {
	Requests = append(Requests, new(ListWitnessesRequest))
}

func (l *ListWitnessesRequest) Load() {
	if models.NodeList == nil && models.NodeList.Addresses == nil {
		panic("list witnesses request load() error")
	}

	if l.Parameters == nil {
		l.Parameters = make([]*Parameter, 0)
	}

	for _, node := range models.NodeList.Addresses {
		param := new(Parameter)
		param.RequestUrl = fmt.Sprintf(
			urlTemplateListWitnesses,
			node.Ip,
			node.HttpPort,
			config.NewNodeType(node.Type).GetApiPathByNodeType())
		param.Node = fmt.Sprintf("%s:%d", node.Ip, node.HttpPort)
		param.Type = node.Type
		param.Tag = node.Tag

		l.Parameters = append(l.Parameters, param)
	}

	logs.Info(
		"list witnesses request load() success, node size:",
		len(l.Parameters),
	)
}

func (l *ListWitnessesRequest) Request() {
	if l.Parameters == nil || len(l.Parameters) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(l.Parameters))
	for _, param := range l.Parameters {
		go l.request(param, &wg)
	}

	wg.Wait()
}

func (l *ListWitnessesRequest) Save2db() {

}

func (l *ListWitnessesRequest) request(param *Parameter, wg *sync.WaitGroup) {
	defer wg.Done()

	response, err := http.Get(param.RequestUrl)

	if err != nil {
		logs.Debug("(", param.RequestUrl, ")", "[http get]", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		logs.Warn("list witnesses request (", param.RequestUrl,
			") response status code",
			response.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Warn("(", param.RequestUrl, ") ", "[read body]", err)
		return
	}

	var witnesses WitnessList
	err = json.Unmarshal(body, &witnesses)

	if err != nil {
		logs.Warn("(", param.RequestUrl, ") ", "[json unmarshal]", err)
		return
	}

	// Report witness
	t := time.Now()
	if witnesses.Witnesses != nil {
		for _, w := range witnesses.Witnesses {
			witnessTags := map[string]string{
				influxDBFieldListWitnessesNode: param.Node,
				influxDBFieldListWitnessesType: param.Type,
				influxDBFieldListWitnessesTag:  param.Tag,
				influxDBFieldListWitnessesUrl:  w.Url,
			}

			witnessFields := map[string]interface{}{
				influxDBFieldListWitnessesNode: param.Node,
				influxDBFieldListWitnessesType: param.Type,
				influxDBFieldListWitnessesTag:  param.Tag,

				influxDBFieldListWitnessesTotalMissed: w.TotalMissed,
				influxDBFieldListWitnessesUrl:         w.Url,
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
