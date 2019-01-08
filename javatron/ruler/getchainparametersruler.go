package ruler

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

// 比较每个参数是否变化
// 与之前的值不一样提醒一次
func ChainParametersChangeRuler(
	db *influxdb.InfluxDB,
	t time.Time,
	nodeIp string,
	nodePort int,
	tagName, nodeType string) (*result.Result,
	error) {
	logs.Debug("ChainParametersChangeRuler ruling")

	res := new(result.Result)
	res.Type = 1
	res.Data = make([]result.Data, 0)

	// 获取最新值

	// 获取之前的值

	if len(res.Data) == 0 {
		return nil, nil
	}

	return res, nil
}
