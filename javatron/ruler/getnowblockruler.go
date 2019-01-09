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

var nowBlockUpdateRulerMark = make(map[string]*nowBlockUpdateRulerMarkInfo)

type nowBlockUpdateRulerMarkInfo struct {
	StartTime time.Time
}

// 块更新规则：
// 1.获取所有节点最大块号maxBlockNum
// 2.获取节点最大块号maxNodeBlockNum
// 3.maxBlockNum - maxNodeBlockNum > 100报警
func NowBlockUpdateRuler(db *influxdb.InfluxDB,
	t time.Time,
	nodeIp string,
	nodePort int,
	tagName, nodeType string) (*result.Result,
	error) {
	logs.Debug("TotalMissedRule ruling")

	res := new(result.Result)
	res.Type = result.NowBlockUpdate
	res.Data = make([]result.Data, 0)

	maxBlockNum, err := getMaxBlockNumByTag(db, tagName, t)
	if err != nil {
		return nil, err
	}

	maxNodeBlockNum, err := getMaxBlockNumByNode(db, nodeIp, nodePort, t)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s:%d", nodeIp, nodePort)

	if maxBlockNum-maxNodeBlockNum > 100 {
		if _, ok := nowBlockUpdateRulerMark[key]; ok {
			return nil, nil
		}

		nowBlockUpdateRulerMark[key] = &nowBlockUpdateRulerMarkInfo{
			StartTime: time.Now(),
		}

		r := &result.NowBlockData{
			Ip:               nodeIp,
			Port:             nodePort,
			Type:             "",
			TagName:          tagName,
			BlockNum:         maxNodeBlockNum,
			ExpectedBlockNum: maxBlockNum,
		}

		res.Data = append(res.Data, r)
	} else {
		if markInfo, ok := nowBlockUpdateRulerMark[key]; ok {
			r := &result.RecoveryData{
				Msg:      fmt.Sprintf("%s块更新恢复正常", key),
				Duration: time.Now().Sub(markInfo.StartTime),
			}

			delete(nowBlockUpdateRulerMark, key)

			res.Data = append(res.Data, r)
		}
	}

	if len(res.Data) == 0 {
		return nil, nil
	}

	return res, nil
}

// 获取分组最大的号
func getMaxBlockNumByTag(
	db *influxdb.InfluxDB,
	tagName string,
	t time.Time) (int64, error) {
	q := fmt.Sprintf(
		`SELECT max(Number) FROM api_get_now_block WHERE time <= %s AND time >= %s AND
"TagName" = '%s'`,
		fmt.Sprintf("%dms", t.UnixNano()/1000000),
		fmt.Sprintf("%dms", t.UnixNano()/1000000-internal10min),
		tagName)

	return getMaxBlock(db, q)
}

func getMaxBlockNumByNode(db *influxdb.InfluxDB,
	nodeIp string,
	nodePort int,
	t time.Time) (int64, error) {
	q := fmt.Sprintf(
		`SELECT max(Number) FROM api_get_now_block WHERE time <= %s AND time >= %s AND
"Node" = '%s:%d'`,
		fmt.Sprintf("%dms", t.UnixNano()/1000000),
		fmt.Sprintf("%dms", t.UnixNano()/1000000-internal10min),
		nodeIp,
		nodePort)

	return getMaxBlock(db, q)
}

func getMaxBlock(db *influxdb.InfluxDB, q string) (int64, error) {
	res, err := db.QueryDB(q)
	if err != nil {
		return 0, err
	}

	if res == nil || len(res) == 0 ||
		res[0].Series == nil || len(res[0].Series) == 0 ||
		res[0].Series[0].Values == nil || len(res[0].Series[0].Values) < 1 {
		return 0, errors.New("no data")
	}

	val := res[0].Series[0].Values[0][1].(json.Number)

	v, err := val.Int64()
	if err != nil {
		return 0, err
	}

	return v, nil
}
