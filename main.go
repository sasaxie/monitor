package main

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/robfig/cron"
	"github.com/sasaxie/monitor/alerts"
	"github.com/sasaxie/monitor/collector"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/engine"
	"github.com/sasaxie/monitor/fetcher"
	"github.com/sasaxie/monitor/javatron/parser"
	"github.com/sasaxie/monitor/javatron/ruler"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/reports"
	"github.com/sasaxie/monitor/result"
	_ "github.com/sasaxie/monitor/routers"
	"github.com/sasaxie/monitor/sender"
	"github.com/sasaxie/monitor/storage"
	"github.com/sasaxie/monitor/storage/influxdb"
	"strings"
	"time"
)

const urlTemplateGetNowBlock = "http://%s:%d/%s/getnowblock"

var e = engine.NewEngine()

func main() {

	initMonitors()

	var err error
	e.DB, err = influxdb.NewInfluxDB(
		config.MonitorConfig.InfluxDB.Address,
		config.MonitorConfig.InfluxDB.Username,
		config.MonitorConfig.InfluxDB.Password)
	if err != nil {
		panic(err)
	}

	t := time.Tick(time.Minute)

	for {
		select {
		case <-t:
			e.Run()
		}
	}
}

func initMonitors() {
	for _, node := range models.NodeList.Addresses {
		if strings.Contains(node.Monitor, "NowBlock") {
			monitor := &engine.Monitor{
				Url: fmt.Sprintf(
					urlTemplateGetNowBlock,
					node.Ip,
					node.HttpPort,
					config.NewNodeType(node.Type).GetApiPathByNodeType()),
				Node: &engine.Node{
					IP:   node.Ip,
					Port: node.HttpPort,
					Tag:  node.TagName,
					Type: node.Type,
				},
				Fetcher: fetcher.DefaultFetcher,
				Parser:  parser.GetNowBlockParser,
				Storage: storage.GetNowBlockStorage,
				Rulers: []func(
					db *influxdb.InfluxDB,
					t time.Time,
					nodeIp string,
					nodePort int,
					tagName, nodeType string) (*result.Result, error){
					ruler.NowBlockUpdateRuler,
				},
				Senders: []func(res ...result.Result) error{
					sender.NilSend,
				},
			}

			e.AddMonitor(monitor)
		}
	}

	e.AddMonitor(&engine.Monitor{
		Url: "http://54.236.37.243:8090/wallet/getchainparameters",
		Node: &engine.Node{
			IP:   "",
			Port: 0,
			Tag:  "",
			Type: "",
		},
		Fetcher: fetcher.NilFetcher,
		Parser:  parser.NilParser,
		Storage: storage.NilStorage,
		Rulers: []func(
			db *influxdb.InfluxDB,
			t time.Time,
			nodeIp string,
			nodePort int,
			tagName, nodeType string) (*result.Result,
			error){
			ruler.NilRule,
		},
		Senders: []func(res ...result.Result) error{
			sender.NilSend,
		},
	})

	e.AddMonitor(&engine.Monitor{
		Url: "http://127.0.0.1:8090/wallet/listwitnesses",
		Node: &engine.Node{
			IP:   "",
			Port: 0,
			Tag:  "",
			Type: "",
		},
		Fetcher: fetcher.DefaultFetcher,
		Parser:  parser.ListWitnessesParser,
		Storage: storage.ListWitnessesStorage,
		Rulers: []func(
			db *influxdb.InfluxDB,
			t time.Time,
			nodeIp string,
			nodePort int,
			tagName, nodeType string) (*result.Result,
			error){
			ruler.TotalMissedRuler,
			ruler.WitnessChangeRuler,
		},
		Senders: []func(res ...result.Result) error{
			sender.NilSend,
		},
	})

	logs.Debug(fmt.Sprintf("init monitors count: %d", len(e.Monitors)))
}

func report() {
	c := cron.New()
	c.AddFunc("0 0 11 * * *", func() {
		logs.Debug("report start")
		r := new(reports.TotalMissed)
		r.Date = time.Now().AddDate(0, 0, -1)
		logs.Debug("report date", r.Date.Format("2006-01-02 15:04:05"))
		r.ComputeData()
		r.Save()
		r.Report()
	})
	c.Start()
}

func change() {
	c := new(alerts.ChainParameters)
	c.MonitorUrl = config.MonitorConfig.Task.ProposalsMonitorUrl
	logs.Info("init proposals monitor url:", c.MonitorUrl)

	ticker := time.NewTicker(
		time.Duration(config.MonitorConfig.Task.GetDataInterval) *
			time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.RequestData()
			c.Judge()
		}
	}
}

func start() {
	getNowBlockAlert := new(alerts.GetNowBlockAlert)
	getNowBlockAlert.Load()

	listWitnessAlert := new(alerts.ListWitnessesAlert)
	listWitnessAlert.Load()

	ticker := time.NewTicker(
		time.Duration(config.MonitorConfig.Task.GetDataInterval) *
			time.Second)
	defer ticker.Stop()

	startAlertCount := 0
	alertFinish := true

	for {
		select {
		case <-ticker.C:
			logs.Debug("start")

			for _, r := range collector.Collectors {
				go r.Collect()
			}

			time.Sleep(10 * time.Second)
			startAlertCount++

			if startAlertCount > 10 && alertFinish {
				alertFinish = false
				getNowBlockAlert.Start()
				getNowBlockAlert.Alert()

				listWitnessAlert.Start()
				listWitnessAlert.Alert()
				alertFinish = true
			}
		}
	}
}
