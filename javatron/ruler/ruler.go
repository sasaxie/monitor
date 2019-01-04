package ruler

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
	"github.com/sasaxie/monitor/storage/influxdb"
)

func NilRule(db *influxdb.InfluxDB) (result.Result, error) {
	logs.Debug("nil ruling")
	return result.Result{}, nil
}
