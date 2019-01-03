package alerts

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/models"
	"github.com/sasaxie/monitor/senders/message"
	"strings"
	"time"
)

type ListWitnessesAlert struct {
	Nodes                 []*Node
	TotalMissedResult     map[string]*ListWitnessesAlertTotalMissedMsg
	WitnessesChangeResult *ListWitnessesAlertWitnessesChangeMsg
	Witnesses1            map[string]bool
	Witnesses2            map[string]bool
	TotalMissed1          map[string]*TotalMissedInfo
	TotalMissed2          map[string]*TotalMissedInfo
}

type TotalMissedInfo struct {
	Address     string
	Url         string
	TagName     string
	TotalMissed int64
}

type ListWitnessesAlertTotalMissedMsg struct {
	Address      string
	Url          string
	TagName      string
	TotalMissed1 int64
	TotalMissed2 int64

	StartTime time.Time
	FreshTime time.Time
	IsFresh   bool
	IsRecover bool
	Msg       string
}

type ListWitnessesAlertWitnessesChangeMsg struct {
	OldWitnesses []*WitnessesChangeInfo
	NewWitnesses []*WitnessesChangeInfo
	Msg          string
}

type WitnessesChangeInfo struct {
	Address string
	Url     string
	TagName string
}

func (l ListWitnessesAlertTotalMissedMsg) TotalMissedChangeString() string {
	return fmt.Sprintf(`address: %s
url: %s
totalMissed: [%d] -> [%d]
msg: %s`, l.Address, l.Url, l.TotalMissed1, l.TotalMissed2, l.Msg)
}

func (l ListWitnessesAlertWitnessesChangeMsg) WitnessChangeString() string {
	res := ""

	for i, v := range l.OldWitnesses {
		if i == 0 {
			res += "SR列表更新\n"
			res += "旧SR\n"
		}
		res += fmt.Sprintf("[%s, %s]\n", v.Url, v.Address)
	}

	for i, v := range l.NewWitnesses {
		if i == 0 {
			if len(res) == 0 && strings.EqualFold(res, "") {
				res += "SR列表有变化\n"
			}

			res += "\n"
			res += "新SR\n"
		}

		res += fmt.Sprintf("[%s, %s]\n", v.Url, v.Address)
	}

	return res
}

func (l *ListWitnessesAlert) Load() {
	if models.NodeList == nil && models.NodeList.Addresses == nil {
		panic("get now block alert load() error")
	}

	if l.Nodes == nil {
		l.Nodes = make([]*Node, 0)
	}

	for _, node := range models.NodeList.Addresses {
		if strings.Contains(node.Monitor, "BlockMissed") {
			n := new(Node)
			n.Ip = node.Ip
			n.GrpcPort = node.GrpcPort
			n.HttpPort = node.HttpPort
			n.Type = node.Type
			n.TagName = node.TagName

			l.Nodes = append(l.Nodes, n)
		}
	}

	logs.Info(
		"list witnesses alert load() success, node size:",
		len(l.Nodes))
}

/**
 Rules:
	1. TotalMissed changed
*/
func (l *ListWitnessesAlert) Start() {
	l.TotalMissedResult = make(map[string]*ListWitnessesAlertTotalMissedMsg)
	l.WitnessesChangeResult = new(ListWitnessesAlertWitnessesChangeMsg)
	l.WitnessesChangeResult.OldWitnesses = make([]*WitnessesChangeInfo, 0)
	l.WitnessesChangeResult.NewWitnesses = make([]*WitnessesChangeInfo, 0)
	l.TotalMissed1 = make(map[string]*TotalMissedInfo)
	l.TotalMissed2 = make(map[string]*TotalMissedInfo)
	l.Witnesses1 = make(map[string]bool)
	l.Witnesses2 = make(map[string]bool)

	t := time.Now().UnixNano() / 1000000

	l.updateWitnesses(t)

	if len(l.Witnesses1) != 0 && len(l.Witnesses2) != 0 {
		for k := range l.Witnesses1 {
			if _, ok := l.Witnesses2[k]; ok {
				l.Witnesses1[k] = true
				l.Witnesses2[k] = true
			}
		}
	}

	for k, v := range l.Witnesses1 {
		if !v {
			u, _ := l.getWitnessUrl(k, t)

			l.WitnessesChangeResult.OldWitnesses = append(l.
				WitnessesChangeResult.OldWitnesses, &WitnessesChangeInfo{
				Address: k,
				Url:     u,
			})
			l.WitnessesChangeResult.Msg = "SR改变"
		}
	}

	for k, v := range l.Witnesses2 {
		u, _ := l.getWitnessUrl(k, t)

		if !v {
			l.WitnessesChangeResult.NewWitnesses = append(l.
				WitnessesChangeResult.NewWitnesses, &WitnessesChangeInfo{
				Address: k,
				Url:     u,
			})
			l.WitnessesChangeResult.Msg = "SR改变"
		}
	}

	l.updateTotalMissed(t)

	for k, v := range l.TotalMissed1 {
		vv := l.TotalMissed2[k]

		if v.TotalMissed != vv.TotalMissed {
			l.TotalMissedResult[k] = &ListWitnessesAlertTotalMissedMsg{
				Address:      k,
				Url:          v.Url,
				TagName:      v.TagName,
				TotalMissed1: v.TotalMissed,
				TotalMissed2: vv.TotalMissed,
				Msg:          "出块超时",
			}
		}
	}
}

func (l *ListWitnessesAlert) updateTotalMissed(t int64) {
	for a, isWitness := range l.Witnesses2 {
		if isWitness {
			totalMissed, u := l.getTotalMissedInfo(a, t-Internal1min)
			l.TotalMissed1[a] = &TotalMissedInfo{
				TotalMissed: totalMissed,
				Url:         u,
				Address:     a,
			}

			totalMissed2, u2 := l.getTotalMissedInfo(a, t)
			l.TotalMissed2[a] = &TotalMissedInfo{
				TotalMissed: totalMissed2,
				Url:         u2,
				Address:     a,
			}
		}
	}
}

func (l *ListWitnessesAlert) getTotalMissedInfo(a string, t int64) (int64,
	string) {
	totalMissed, _ := l.getTotalMissed(a, t)
	u, _ := l.getWitnessUrl(a, t)

	return totalMissed, u
}

func (l *ListWitnessesAlert) getWitnessUrl(a string, t int64) (string,
	error) {
	q := fmt.Sprintf(`SELECT Url FROM api_list_witnesses WHERE
Address='%s' AND time <= %s AND time > %s ORDER BY time DESC LIMIT 1`, a,
		fmt.Sprintf("%dms", t),
		fmt.Sprintf("%dms", t-Internal5min))

	res, err := influxdb.QueryDB(influxdb.Client.C, q)
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

func (l *ListWitnessesAlert) getTotalMissed(a string, t int64) (int64,
	error) {
	q := fmt.Sprintf(`SELECT max(TotalMissed) FROM api_list_witnesses WHERE
Address='%s' AND time <= %s AND time > %s`, a,
		fmt.Sprintf("%dms", t),
		fmt.Sprintf("%dms", t-Internal5min))

	res, err := influxdb.QueryDB(influxdb.Client.C, q)
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

func (l *ListWitnessesAlert) updateWitnesses(t int64) error {

	l.getWitness1(t - Internal1min)
	l.getWitness2(t)

	return nil
}

func (l *ListWitnessesAlert) getWitness1(t int64) error {
	q := fmt.Sprintf(`SELECT distinct(
Address) FROM api_list_witnesses WHERE IsJobs=true AND time <= %s AND time
>= %s`, fmt.Sprintf("%dms", t), fmt.Sprintf("%dms",
		t-Internal5min))

	res, err := influxdb.QueryDB(influxdb.Client.C, q)
	if err != nil {
		return err
	}

	if len(res) == 0 ||
		len(res[0].Series) == 0 ||
		len(res[0].Series[0].Values) == 0 {
		return errors.New("no data")
	}

	for _, val := range res[0].Series[0].Values {
		address := val[1].(string)

		l.Witnesses1[address] = false
	}

	return nil
}

func (l *ListWitnessesAlert) getWitness2(t int64) error {
	q := fmt.Sprintf(`SELECT distinct(
Address) FROM api_list_witnesses WHERE IsJobs=true AND time <= %s AND time
>= %s`, fmt.Sprintf("%dms", t), fmt.Sprintf("%dms", t-Internal5min))

	res, err := influxdb.QueryDB(influxdb.Client.C, q)
	if err != nil {
		return err
	}

	if len(res) == 0 ||
		len(res[0].Series) == 0 ||
		len(res[0].Series[0].Values) == 0 {
		return errors.New("no data")
	}

	for _, val := range res[0].Series[0].Values {
		address := val[1].(string)

		l.Witnesses2[address] = false
	}

	return nil
}

func (l *ListWitnessesAlert) Alert() {
	for _, v := range l.TotalMissedResult {
		config.Ding.Send(message.Alert, v.TotalMissedChangeString())
	}

	res := l.WitnessesChangeResult.WitnessChangeString()
	if len(res) > 0 && !strings.EqualFold(res, "") {
		msg := fmt.Sprintf(`
			{
				"msgtype": "text",
				"text": {
					"content": "%s"
				}
			}
			`, res)

		config.Ding.Send(message.Alert, msg)
	}
}
