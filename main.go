package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/alerts"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/datamanger"
	_ "github.com/sasaxie/monitor/routers"
	"time"
)

func main() {

	logs.Info("start monitor")

	go start()

	defer influxdb.Client.C.Close()

	beego.Run()
}

func start() {
	for _, r := range datamanger.Requests {
		r.Load()
	}

	a := new(alerts.GetNowBlockAlert)
	a.Load()

	b := new(alerts.ListWitnessesAlert)
	b.Load()

	ticker := time.NewTicker(
		time.Duration(config.MonitorConfig.Task.GetDataInterval) *
			time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logs.Debug("start")

			for _, r := range datamanger.Requests {
				go r.Request()
			}

			time.Sleep(10 * time.Second)

			a.Start()
			a.Alert()

			b.Start()
			b.Alert()
		}
	}
}
