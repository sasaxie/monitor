package engine

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

type Engine struct {
	Monitors []*Monitor

	DB *influxdb.InfluxDB
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

	Rulers []func(db *influxdb.InfluxDB, t time.Time) (*result.Result, error)

	Senders []func(results ...result.Result) error
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
			logs.Error(err)
			continue
		}

		parseData, err := monitor.Parser(data)
		if err != nil {
			logs.Error(err)
			continue
		}

		err = monitor.Storage(
			e.DB,
			parseData,
			"", "", "")

		if err != nil {
			logs.Error(err)
			continue
		}

		t := time.Now()

		results := make([]*result.Result, 0)
		for _, r := range monitor.Rulers {
			res, err := r(e.DB, t)
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
			logs.Debug("结果类型：", res.Type, len(res.Data))
			for _, d := range res.Data {
				logs.Debug(d.ToMsg())
			}
		}
	}
}
