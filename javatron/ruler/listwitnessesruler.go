package ruler

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
	"github.com/sasaxie/monitor/storage/influxdb"
)

// 出块超时的规则：
// 1.获取最近10分钟的所有Witness
// 2.获取这些Witness的最近10分钟内的最小TotalMissed和最大TotalMissed
// 3.如果最小值和最大值不一样就报警
func TotalMissedRule(db *influxdb.InfluxDB) (result.Result, error) {
	logs.Debug("TotalMissedRule ruling")

	return result.Result{}, nil
}
