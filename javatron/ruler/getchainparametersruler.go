package ruler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
	"github.com/sasaxie/monitor/storage/influxdb"
	"time"
)

const SpecialChainParameterValue int64 = -123123123

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
	res.Type = result.ChainParametersChange
	res.Data = make([]result.Data, 0)

	// 获取最新值
	currentParameters, err := getChainParameters(db, t)
	if err != nil {
		return nil, err
	}

	// 获取之前的值
	previousParameters, err := getChainParameters(db, t.Add(-time.Minute))
	if err != nil {
		return nil, err
	}

	for key, newValue := range currentParameters {
		if oldValue, ok := previousParameters[key]; ok {
			if newValue != oldValue &&
				newValue != SpecialChainParameterValue &&
				oldValue != SpecialChainParameterValue {
				r := &result.ChainParametersChangeData{
					Key:      key,
					OldValue: oldValue,
					NewValue: newValue,
				}

				res.Data = append(res.Data, r)
			}
		}
	}

	if len(res.Data) == 0 {
		return nil, nil
	}

	return res, nil
}

func getChainParameters(
	db *influxdb.InfluxDB,
	t time.Time) (map[string]int64, error) {
	q := fmt.Sprintf(`
SELECT * FROM api_chain_parameters WHERE time <= %s AND time >= %s ORDER BY
time DESC LIMIT 1`,
		fmt.Sprintf("%dms", t.UnixNano()/1000000),
		fmt.Sprintf("%dms", t.UnixNano()/1000000-internal10min))

	res, err := db.QueryDB(q)

	if err != nil {
		return nil, err
	}

	if res == nil || len(res) == 0 ||
		res[0].Series == nil || len(res[0].Series) == 0 ||
		res[0].Series[0].Values == nil || len(res[0].Series[0].Values) < 1 {
		return nil, errors.New("no data")
	}

	chainParameters := make(map[string]int64)
	for _, val := range res[0].Series[0].Values {
		var tmpValue int64 = SpecialChainParameterValue
		for i := 1; i < len(val); i++ {
			if i%2 != 0 {
				tmpValue, err = val[i].(json.Number).Int64()
				if err != nil {
					logs.Error(err)
					tmpValue = SpecialChainParameterValue
				}
			} else {
				chainParameters[val[i].(string)] = tmpValue
			}
		}
	}

	return chainParameters, nil
}
