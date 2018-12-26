package alerts

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/database/influxdb"
	"github.com/sasaxie/monitor/dingding"
	"github.com/sasaxie/monitor/models"
	"strings"
	"time"
)

// ms: 1min
const internal1min int64 = 1000 * 60 * 1

type ListWitnessesAlert struct {
	Nodes                 []*Node
	TotalMissedResult     map[string]*ListWitnessesAlertTotalMissedMsg
	WitnessesChangeResult *ListWitnessesAlertWitnessesChangeMsg
	Witnesses1            map[string]bool
	Witnesses2            map[string]bool
	TotalMissed1          map[string]int64
	TotalMissed2          map[string]int64
}

type ListWitnessesAlertTotalMissedMsg struct {
	Address      string
	Url          string
	TotalMissed1 int64
	TotalMissed2 int64

	StartTime time.Time
	FreshTime time.Time
	IsFresh   bool
	IsRecover bool
	Msg       string
}

type ListWitnessesAlertWitnessesChangeMsg struct {
	OldWitnesses []string
	NewWitnesses []string
	Msg          string
}

func (l ListWitnessesAlertTotalMissedMsg) TotalMissedChangeString() string {
	return fmt.Sprintf(`address: %s
totalMissed: [%d] -> [%d]
msg: %s`, l.Address, l.TotalMissed1, l.TotalMissed2, l.Msg)
}

func (l ListWitnessesAlertWitnessesChangeMsg) WitnessChangeString() string {
	res := ""

	for i, v := range l.OldWitnesses {
		if i == 0 {
			res += "SR列表更新\n"
			res += "旧SR\n"
		}
		res += v + "\n"
	}

	for i, v := range l.NewWitnesses {
		if i == 0 {
			if len(res) == 0 && strings.EqualFold(res, "") {
				res += "SR列表有变化\n"
			}

			res += "\n"
			res += "新SR\n"
		}

		res += v + "\n"
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
			n.Tag = node.Tag

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
	l.WitnessesChangeResult.OldWitnesses = make([]string, 0)
	l.WitnessesChangeResult.NewWitnesses = make([]string, 0)
	l.TotalMissed1 = make(map[string]int64)
	l.TotalMissed2 = make(map[string]int64)
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
			l.WitnessesChangeResult.OldWitnesses = append(l.
				WitnessesChangeResult.OldWitnesses, k)
			l.WitnessesChangeResult.Msg = "SR改变"
		}
	}

	for k, v := range l.Witnesses2 {
		if !v {
			l.WitnessesChangeResult.NewWitnesses = append(l.
				WitnessesChangeResult.NewWitnesses, k)
			l.WitnessesChangeResult.Msg = "SR改变"
		}
	}

	l.updateTotalMissed(t)

	for k, v := range l.TotalMissed1 {
		vv := l.TotalMissed2[k]

		if v != vv {
			l.TotalMissedResult[k] = &ListWitnessesAlertTotalMissedMsg{
				Address:      k,
				TotalMissed1: v,
				TotalMissed2: vv,
				Msg:          "出块超时",
			}
		}
	}

	logs.Debug("list witnesses alert finished")
}

func (l *ListWitnessesAlert) updateTotalMissed(t int64) {
	for a, isWitness := range l.Witnesses2 {
		if isWitness {
			l.TotalMissed1[a], _ = l.getTotalMissed(a, t-internal1min)
			l.TotalMissed2[a], _ = l.getTotalMissed(a, t)
		}
	}
}

func (l *ListWitnessesAlert) getTotalMissed(a string, t int64) (int64,
	error) {
	q := fmt.Sprintf(`SELECT max(TotalMissed) FROM api_list_witnesses WHERE
Address='%s' AND time <= %s AND time > %s`, a,
		fmt.Sprintf("%dms", t),
		fmt.Sprintf("%dms", t-internal5min))

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

	l.getWitness1(t - internal1min)
	l.getWitness2(t)

	return nil
}

func (l *ListWitnessesAlert) getWitness1(t int64) error {
	q := fmt.Sprintf(`SELECT distinct(
Address) FROM api_list_witnesses WHERE IsJobs=true AND time <= %s AND time
>= %s`, fmt.Sprintf("%dms", t), fmt.Sprintf("%dms",
		t-internal5min))

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
>= %s`, fmt.Sprintf("%dms", t), fmt.Sprintf("%dms", t-internal5min))

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
		bodyContent := fmt.Sprintf(`
			{
				"msgtype": "text",
				"text": {
					"content": "%s"
				}
			}
			`, v.TotalMissedChangeString())

		dingding.DingAlarm.Alarm([]byte(bodyContent))
	}

	res := l.WitnessesChangeResult.WitnessChangeString()
	if len(res) > 0 && !strings.EqualFold(res, "") {
		bodyContent := fmt.Sprintf(`
			{
				"msgtype": "text",
				"text": {
					"content": "%s"
				}
			}
			`, res)

		dingding.DingAlarm.Alarm([]byte(bodyContent))
	}
}
