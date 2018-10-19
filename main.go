package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/database/influxdb"
	_ "github.com/sasaxie/monitor/routers"
	"github.com/sasaxie/monitor/task"
)

func main() {

	logs.Info("start monitor")

	go task.StartGrpcMonitor()
	go task.StartHttpMonitor()

	defer influxdb.Client.C.Close()

	beego.Run()
}
