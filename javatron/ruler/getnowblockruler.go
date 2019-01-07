package ruler

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

// 块更新规则：
// 1.获取所有节点最大块号m1
// 2.获取节点最大块号m2
// 3.m1 - m2 > 100报警
// 4.m1与之前没变化报警
func NowBlockUpdateRuler(db *influxdb.InfluxDB,
	t time.Time) (*result.Result,
	error) {
	logs.Debug("TotalMissedRule ruling")

	res := new(result.Result)
	res.Type = 1
	res.Data = make([]result.Data, 0)

	// 获取m1

	// 获取每个节点的m2，并与m1比较

	if len(res.Data) == 0 {
		return nil, nil
	}

	return res, nil
}
