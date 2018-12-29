package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/robfig/cron"
	"github.com/sasaxie/monitor/alerts"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/datamanger"
	"github.com/sasaxie/monitor/reports"
	_ "github.com/sasaxie/monitor/routers"
	"time"
)

func main() {

	logs.Info("start monitor")

	go start()
	go report()
	go change()

	defer influxdb.Client.C.Close()

	beego.Run()
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
	for _, r := range datamanger.Requests {
		r.Load()
	}

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

			for _, r := range datamanger.Requests {
				go r.Request()
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
