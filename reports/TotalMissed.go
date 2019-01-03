package reports

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/senders/message"
	"time"
)

const (
	influxDBPointNameReportTotalMissed    = "report_total_missed"
	influxDBPointNameReportTotalMissedAll = "report_total_missed_all"
	influxDBTagAddress                    = "Address"
	influxDBTagUrl                        = "Url"
	influxDBTagTotalMissed                = "TotalMissed"
	influxDBTagTag                        = "TagName"
	influxDBTagDate                       = "Date"
	influxDBTagPercent                    = "Percent"
	influxDBTagAll                        = "All"
)

type TotalMissed struct {
	Data map[string]*OriginData
	Date time.Time
}

type OriginData struct {
	TotalMissed int64
	Url         string
}

func (t *TotalMissed) addData(k string, v *OriginData) {
	if t.Data == nil {
		t.Data = make(map[string]*OriginData)
	}

	t.Data[k] = v
}

func (t *TotalMissed) getT1T2() (int64, int64) {
	startTime := time.Date(
		t.Date.Year(),
		t.Date.Month(),
		t.Date.Day(),
		0,
		0,
		0,
		0,
		time.Local)

	endTime := time.Date(
		t.Date.Year(),
		t.Date.Month(),
		t.Date.Day(),
		23,
		59,
		59,
		99999999,
		time.Local)

	logs.Debug(
		"report total missed start time:",
		startTime.Format("2006-01-02 15:04:05"),
		", end time:",
		endTime.Format("2006-01-02 15:04:05"))

	t1, t2 := startTime.UnixNano()/1000000, endTime.UnixNano()/1000000

	return t1, t2
}

// Get data from InfluxDB range t1 and t2.
func (t *TotalMissed) ComputeData() {
	// 获取这个时间段的所有的不同的address
	addresses := t.getAllAddress()
	if addresses == nil || len(addresses) == 0 {
		return
	}

	// 获取每个address的总的totalMissed
	for _, address := range addresses {
		sum := t.getTotalMissedSum(address)
		if sum == 0 {
			continue
		}
		t.addData(address, &OriginData{TotalMissed: sum})
	}

	if t.Data == nil {
		return
	}
	for address, originData := range t.Data {
		originData.Url = t.getUrlByAddress(address)
	}
}

func (t *TotalMissed) getTotalMissedSum(address string) int64 {
	t1, t2 := t.getT1T2()

	endQuery := fmt.Sprintf(`
SELECT max(TotalMissed) FROM api_list_witnesses WHERE time >= %s AND time
<= %s AND TagName='主网' AND Address='%s'
`, fmt.Sprintf("%dms", t1), fmt.Sprintf("%dms", t2), address)

	startQuery := fmt.Sprintf(`
SELECT min(TotalMissed) FROM api_list_witnesses WHERE time >= %s AND time
<= %s AND TagName='主网' AND Address='%s'
`, fmt.Sprintf("%dms", t1), fmt.Sprintf("%dms", t2), address)

	startTotalMissed := t.getTotalMissed(startQuery)
	endTotalMissed := t.getTotalMissed(endQuery)

	logs.Debug("max total missed:", endTotalMissed, "min total missed:",
		startTotalMissed)
	return endTotalMissed - startTotalMissed
}

func (t *TotalMissed) getTotalMissed(q string) int64 {
	res, err := influxdb.QueryDB(influxdb.Client.C, q)
	if err != nil {
		logs.Error("[package: reports] [method: getTotalMissed("+
			")] [QueryDB error]", err.Error())
		return 0
	}

	if res == nil || len(res) == 0 ||
		len(res[0].Series) == 0 ||
		len(res[0].Series[0].Values) == 0 {
		return 0
	}

	var totalMissed int64 = 0
	for _, value := range res[0].Series[0].Values {
		totalMissed, err = value[1].(json.Number).Int64()
	}

	return totalMissed
}

func (t *TotalMissed) getAllAddress() []string {
	t1, t2 := t.getT1T2()

	q := fmt.Sprintf(`
SELECT distinct(Address) FROM api_list_witnesses WHERE time >= %s AND time
<= %s AND TagName='主网' AND IsJobs=true
`, fmt.Sprintf("%dms", t1), fmt.Sprintf("%dms", t2))

	res, err := influxdb.QueryDB(influxdb.Client.C, q)
	if err != nil {
		logs.Error("[package: reports] [method: getAllAddress("+
			")] [QueryDB error]", err.Error())
		return nil
	}

	addresses := make([]string, 0)
	if res == nil || len(res) == 0 ||
		len(res[0].Series) == 0 ||
		len(res[0].Series[0].Values) == 0 {
		return nil
	}
	for _, value := range res[0].Series[0].Values {
		addresses = append(addresses, value[1].(string))
	}

	logs.Debug(fmt.Sprintf("[package: reports] [method: getAllAddress("+
		")] [address len: %d]", len(addresses)))

	return addresses
}

func (t *TotalMissed) Save() {
	if t.Data == nil {
		return
	}

	reportTime := time.Now()

	var totalMissedSum int64 = 0
	for address, originData := range t.Data {
		tags := map[string]string{
			influxDBTagAddress: address,
			influxDBTagUrl:     originData.Url,
		}

		fields := map[string]interface{}{
			influxDBTagAddress:     address,
			influxDBTagTag:         "主网",
			influxDBTagTotalMissed: fmt.Sprintf("%d", originData.TotalMissed),
			influxDBTagUrl:         originData.Url,
			influxDBTagDate:        t.Date.UnixNano() / 1000000,
		}

		influxdb.Client.WriteByTime(
			influxDBPointNameReportTotalMissed,
			tags,
			fields,
			reportTime)

		totalMissedSum += originData.TotalMissed
	}

	if totalMissedSum != 0 {
		tags := map[string]string{}

		fields := map[string]interface{}{
			influxDBTagTag:     "主网",
			influxDBTagDate:    t.Date.UnixNano() / 1000000,
			influxDBTagPercent: 1.0 * totalMissedSum / 28800,
			influxDBTagAll:     totalMissedSum,
		}

		influxdb.Client.WriteByTime(
			influxDBPointNameReportTotalMissedAll,
			tags,
			fields,
			reportTime)
	}
}

func (t *TotalMissed) getUrlByAddress(address string) string {
	t1, t2 := t.getT1T2()

	q := fmt.Sprintf(`
SELECT last(Url) FROM api_list_witnesses WHERE time >= %s AND time
<= %s AND TagName='主网' AND Address='%s'
`, fmt.Sprintf("%dms", t1), fmt.Sprintf("%dms", t2), address)

	res, err := influxdb.QueryDB(influxdb.Client.C, q)
	if err != nil {
		logs.Error("[package: reports] [method: getUrlByAddress("+
			")] [QueryDB error]", err.Error())
		return ""
	}

	if res == nil || len(res) == 0 ||
		len(res[0].Series) == 0 ||
		len(res[0].Series[0].Values) == 0 {
		return ""
	}

	u := ""
	for _, value := range res[0].Series[0].Values {
		u = value[1].(string)
	}

	return u
}

func (t *TotalMissed) Report() {
	var percent = 0.0
	var sumTotalMissed int64 = 0

	msg := ""
	if t.Data != nil {
		for _, originData := range t.Data {
			msg += fmt.Sprintf("%6d: %s\n",
				originData.TotalMissed,
				originData.Url)
			sumTotalMissed += originData.TotalMissed
		}
		percent = float64(sumTotalMissed) / 28800 * 100
	}

	msg += fmt.Sprintf("%s 总Miss数：%d，总Miss率：%.4f%%\n",
		t.Date.Format("2006-01-02"),
		sumTotalMissed,
		percent)

	config.Ding.Send(message.Report, msg)
}
