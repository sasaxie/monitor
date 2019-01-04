package storage

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/storage/influxdb"
)

func NilStorage(
	db *influxdb.InfluxDB,
	data interface{},
	nodeHost, nodeTagName, nodeType string) error {
	logs.Debug("nil storing")
	return nil
}
