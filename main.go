package main

import (
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/task"
	"time"
)

func main() {
	go task.StartGrpcMonitor()

	defer influxdb.Client.C.Close()

	for {
		time.Sleep(time.Minute)
	}
}
