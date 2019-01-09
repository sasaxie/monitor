package result

import (
	"fmt"
	"time"
)

type ResType int

const (
	TotalMissed ResType = iota
	WitnessChange
	NowBlockUpdate
	ChainParametersChange
)

type Result struct {
	Type ResType
	Data []Data
}

type Data interface {
	ToMsg() string
}

type TotalMissedData struct {
	WitnessAddress string
	WitnessUrl     string
	MinTotalMissed int64
	MaxTotalMissed int64
	StartTime      time.Time
	EndTime        time.Time
}

func (t TotalMissedData) ToMsg() string {
	return fmt.Sprintf(`出块超时：
address: %s
url: %s
totalMissed: [%d] -> [%d]
timeRange: [%s] ~ [%s]
`,
		t.WitnessAddress,
		t.WitnessUrl,
		t.MinTotalMissed,
		t.MaxTotalMissed,
		t.StartTime.Format("15:04:05"),
		t.EndTime.Format("15:04:05"))
}

type RecoveryData struct {
	Msg      string
	Duration time.Duration
}

func (r RecoveryData) ToMsg() string {
	return fmt.Sprintf(`恢复正常：
msg: %s
duration: %s
`,
		r.Msg,
		r.Duration.String())
}

type WitnessChangeData struct {
	WitnessAddress string
	IsNew          bool
}

func (w WitnessChangeData) ToMsg() string {
	if w.IsNew {
		return fmt.Sprintf(`新的出块Witness：
address: %s
`,
			w.WitnessAddress)
	} else {
		return fmt.Sprintf(`旧的出块Witness：
address: %s
`,
			w.WitnessAddress)
	}

}

type NowBlockData struct {
	Ip               string
	Port             int
	Type             string
	TagName          string
	BlockNum         int64
	ExpectedBlockNum int64
}

func (n NowBlockData) ToMsg() string {
	return fmt.Sprintf(`块更新异常：
ip: %s
port: %d
type: %s
tagName: %s
blockNum: %d
expectedBlockNum: %d`, n.Ip, n.Port, n.Type, n.TagName, n.BlockNum, n.ExpectedBlockNum)
}

type ChainParametersChangeData struct {
	Key      string
	OldValue int64
	NewValue int64
}

func (c ChainParametersChangeData) ToMsg() string {
	return fmt.Sprintf(`提议值更新：
key: %s
value: [%d] -> [%d]
`, c.Key, c.OldValue, c.NewValue)
}
