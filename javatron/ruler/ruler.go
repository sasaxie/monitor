package ruler

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

func NilRule(db *influxdb.InfluxDB, t time.Time) (*result.Result, error) {
	logs.Debug("nil ruling")
	return nil, nil
}
