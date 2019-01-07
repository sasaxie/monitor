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

const (
	totalMissedDuration  = -10 * time.Minute
	totalMissedThreshold = 5
	internal10min        = 10 * 60 * 1000
)

var totalMissedRuleMark = make(map[string]*totalMissedRuleMarkInfo)

type totalMissedRuleMarkInfo struct {
	IsRecover bool
	IsFresh   bool
	FreshTime time.Time
	StartTime time.Time
}

// 出块超时的规则：
// 1.获取最近10分钟的所有Witness
// 2.获取这些Witness的最近10分钟内的最小TotalMissed和最大TotalMissed
// 3.如果最小值和最大值相差太大就报警
func TotalMissedRuler(db *influxdb.InfluxDB, t time.Time) (*result.Result,
	error) {
	logs.Debug("TotalMissedRule ruling")

	res := new(result.Result)
	res.Type = 1
	res.Data = make([]result.Data, 0)

	startTime := t.Add(totalMissedDuration)
	endTime := t

	witnessAddresses, err := getAllWitnessAddresses(db, startTime, endTime)

	if err != nil {
		return nil, err
	}

	for _, witnessAddress := range witnessAddresses {
		min, max, err := getTotalMissedMinAndMax(db, startTime, endTime,
			witnessAddress)
		if err != nil {
			logs.Error(err)
			continue
		}

		if max-min > totalMissedThreshold {
			if markInfo, ok := totalMissedRuleMark[witnessAddress]; ok {
				// 如果有值，看看是否刷新了
				if time.Now().Sub(markInfo.FreshTime) >= time.Hour {
					markInfo.FreshTime = time.Now()
					markInfo.IsFresh = true
				}
			} else {
				// 如果没值，是新增的值
				totalMissedRuleMark[witnessAddress] = &totalMissedRuleMarkInfo{
					IsRecover: false,
					IsFresh:   true,
					FreshTime: endTime,
					StartTime: endTime,
				}
			}

			markInfo := totalMissedRuleMark[witnessAddress]

			u, _ := getWitnessUrlByAddress(db, witnessAddress, endTime)

			r := &result.TotalMissedData{
				WitnessAddress: witnessAddress,
				WitnessUrl:     u,
				MinTotalMissed: min,
				MaxTotalMissed: max,
				StartTime:      startTime,
				EndTime:        endTime,
				Duration:       endTime.Sub(markInfo.StartTime),
			}

			if markInfo.IsFresh {
				markInfo.IsFresh = false
				res.Data = append(res.Data, r)
			}
		} else {
			if markInfo, ok := totalMissedRuleMark[witnessAddress]; ok {
				// 如果有值，说明之前存在异常，现在恢复了
				r := &result.RecoveryData{
					Msg:      fmt.Sprintf("%s出块超时恢复正常", witnessAddress),
					Duration: time.Now().Sub(markInfo.StartTime),
				}

				delete(totalMissedRuleMark, witnessAddress)

				res.Data = append(res.Data, r)
			}
		}
	}

	if len(res.Data) == 0 {
		return nil, nil
	}

	return res, nil
}

func getAllWitnessAddresses(
	db *influxdb.InfluxDB,
	startTime, endTime time.Time) ([]string, error) {
	witnessAddresses := make([]string, 0)

	q := fmt.Sprintf(`SELECT distinct(
Address) FROM api_list_witnesses WHERE IsJobs=true AND time <= %s AND time
>= %s`, fmt.Sprintf("%dms", endTime.UnixNano()/1000000), fmt.Sprintf("%dms",
		startTime.UnixNano()/1000000))

	res, err := db.QueryDB(q)
	if err != nil {
		return witnessAddresses, err
	}

	if len(res) == 0 ||
		len(res[0].Series) == 0 ||
		len(res[0].Series[0].Values) == 0 {
		return witnessAddresses, errors.New("no data")
	}

	for _, val := range res[0].Series[0].Values {
		address := val[1].(string)

		witnessAddresses = append(witnessAddresses, address)
	}

	return witnessAddresses, nil
}

// 获取Witness时间范围内最大差值
func getTotalMissedMinAndMax(
	db *influxdb.InfluxDB,
	startTime, endTime time.Time,
	address string) (int64, int64, error) {

	maxQ := fmt.Sprintf(`SELECT max(TotalMissed) FROM api_list_witnesses WHERE
Address='%s' AND time <= %s AND time > %s`, address,
		fmt.Sprintf("%dms", endTime.UnixNano()/1000000),
		fmt.Sprintf("%dms", startTime.UnixNano()/1000000))

	minQ := fmt.Sprintf(`SELECT min(TotalMissed) FROM api_list_witnesses WHERE
Address='%s' AND time <= %s AND time > %s`, address,
		fmt.Sprintf("%dms", endTime.UnixNano()/1000000),
		fmt.Sprintf("%dms", startTime.UnixNano()/1000000))

	maxTotalMissed, err := getTotalMissed(db, maxQ)
	if err != nil {
		return 0, 0, err
	}

	minTotalMissed, err := getTotalMissed(db, minQ)
	if err != nil {
		return 0, 0, err
	}

	return minTotalMissed, maxTotalMissed, nil
}

func getTotalMissed(db *influxdb.InfluxDB, q string) (int64, error) {
	res, err := db.QueryDB(q)
	if err != nil {
		return 0, err
	}

	if res == nil || len(res) == 0 ||
		res[0].Series == nil || len(res[0].Series) == 0 ||
		res[0].Series[0].Values == nil || len(res[0].Series[0].Values) < 1 {
		return 0, errors.New("get total missed error: no data")
	}

	val := res[0].Series[0].Values[0][1].(json.Number)

	v, err := val.Int64()
	if err != nil {
		return 0, err
	}

	return v, nil
}

func getWitnessUrlByAddress(
	db *influxdb.InfluxDB,
	address string,
	t time.Time) (string,
	error) {
	ti := t.UnixNano() / 1000000
	q := fmt.Sprintf(`SELECT Url FROM api_list_witnesses WHERE
Address='%s' AND time <= %s AND time > %s ORDER BY time DESC LIMIT 1`, address,
		fmt.Sprintf("%dms", ti),
		fmt.Sprintf("%dms", ti-internal10min))

	res, err := db.QueryDB(q)
	if err != nil {
		return "", err
	}

	if res == nil || len(res) == 0 ||
		res[0].Series == nil || len(res[0].Series) == 0 ||
		res[0].Series[0].Values == nil || len(res[0].Series[0].Values) < 1 {
		return "", errors.New("get total missed url error: no data")
	}

	val := res[0].Series[0].Values[0][1].(string)

	return val, nil
}

type WitnessMark struct {
	CurrentContain  bool
	PreviousContain bool
}

// Witness改变报警
// 这个是实时报警，出现就立即提醒
// 获取当前时间范围内的所有信息
// 获取1分钟前时间范围内的所有信息
// 进行比较
func WitnessChangeRuler(db *influxdb.InfluxDB, t time.Time) (*result.Result, error) {
	logs.Debug("WitnessChangeRuler ruling")

	res := new(result.Result)
	res.Type = 2
	res.Data = make([]result.Data, 0)

	currentWitnessAddresses, err := getAllWitnessAddresses(
		db,
		t.Add(-10*time.Minute),
		t)

	if err != nil {
		return nil, err
	}

	logs.Debug(fmt.Sprintf("WitnessChangeRuler got %d current witness",
		len(currentWitnessAddresses)))

	previousWitnessAddresses, err := getAllWitnessAddresses(
		db,
		t.Add(-11*time.Minute),
		t.Add(-time.Minute))

	if err != nil {
		return nil, err
	}

	logs.Debug(fmt.Sprintf("WitnessChangeRuler got %d previous witness",
		len(previousWitnessAddresses)))

	allWitnessesAddresses := make(map[string]*WitnessMark)

	for _, currentWitnessAddress := range currentWitnessAddresses {
		allWitnessesAddresses[currentWitnessAddress] = &WitnessMark{
			CurrentContain:  true,
			PreviousContain: false,
		}
	}

	for _, previousWitnessAddress := range previousWitnessAddresses {
		if mark, ok := allWitnessesAddresses[previousWitnessAddress]; ok {
			mark.PreviousContain = true
		} else {
			allWitnessesAddresses[previousWitnessAddress] = &WitnessMark{
				CurrentContain:  false,
				PreviousContain: true,
			}
		}
	}

	for k, mark := range allWitnessesAddresses {
		if mark.PreviousContain && !mark.CurrentContain {
			res.Data = append(res.Data, &result.WitnessChangeData{
				WitnessAddress: k,
				IsNew:          false,
			})
		} else if mark.CurrentContain && !mark.PreviousContain {
			res.Data = append(res.Data, &result.WitnessChangeData{
				WitnessAddress: k,
				IsNew:          true,
			})
		}
	}

	if len(res.Data) == 0 {
		return nil, nil
	}

	return res, nil
}
