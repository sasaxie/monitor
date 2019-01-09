package engine

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

var monitorUrlMark = make(map[string]int)

type Engine struct {
	Monitors []*Monitor

	DB *influxdb.InfluxDB

	MsgQueue []string
}

type Monitor struct {
	Url  string
	Node *Node

	Fetcher func(url string) ([]byte, error)
	Parser  func(data []byte) (interface{}, error)

	Storage func(
		db *influxdb.InfluxDB,
		data interface{},
		nodeHost, nodeTagName, nodeType string) error

	Rulers []func(
		db *influxdb.InfluxDB,
		t time.Time,
		nodeIp string,
		nodePort int,
		tagName, nodeType string) (*result.Result, error)
}

type Node struct {
	IP   string
	Port int
	Tag  string
	Type string
}

func NewEngine() *Engine {
	return new(Engine)
}

func (e *Engine) AddMonitor(monitor *Monitor) {
	if e.Monitors == nil {
		e.Monitors = make([]*Monitor, 0)
	}

	e.Monitors = append(e.Monitors, monitor)
}

func (e *Engine) Run() {

	for _, monitor := range e.Monitors {
		data, err := monitor.Fetcher(monitor.Url)
		if err != nil {
			if _, ok := monitorUrlMark[monitor.Url]; !ok {
				monitorUrlMark[monitor.Url] = 1
			}
			continue
		}

		if _, ok := monitorUrlMark[monitor.Url]; ok {
			monitorUrlMark[monitor.Url] = 3
		}

		parseData, err := monitor.Parser(data)
		if err != nil {
			logs.Error(err)
			continue
		}

		err = monitor.Storage(
			e.DB,
			parseData,
			fmt.Sprintf("%s:%d", monitor.Node.IP, monitor.Node.Port),
			monitor.Node.Tag,
			monitor.Node.Type)

		if err != nil {
			logs.Error(err)
			continue
		}

		t := time.Now()

		results := make([]*result.Result, 0)
		for _, r := range monitor.Rulers {
			res, err := r(
				e.DB,
				t,
				monitor.Node.IP,
				monitor.Node.Port,
				monitor.Node.Tag,
				monitor.Node.Type)
			if err != nil {
				logs.Error(err)
				continue
			}

			if res != nil {
				results = append(results, res)
			}
		}

		// 将结果字符串放到队列，待发送
		for _, res := range results {
			for _, d := range res.Data {
				e.MsgQueue = append(e.MsgQueue, d.ToMsg())
			}
		}
	}

	for k, v := range monitorUrlMark {
		if v == 1 {
			// 待发送
			e.MsgQueue = append(e.MsgQueue, fmt.Sprintf("获取接口数据失败：%s", k))
			monitorUrlMark[k] = 2
		} else if v == 3 {
			// 待发送
			e.MsgQueue = append(e.MsgQueue, fmt.Sprintf("获取接口数据恢复：%s", k))
			delete(monitorUrlMark, k)
		}
	}
}
