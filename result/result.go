package result

import (
	"fmt"
	"time"
)

type Result struct {
	Type int
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
	Duration       time.Duration
}

func (t TotalMissedData) ToMsg() string {
	return fmt.Sprintf(`出块超时：
address: %s
url: %s
totalMissed: [%d] -> [%d]
timeRange: [%s] ~ [%s]
duration: %s
`,
		t.WitnessAddress,
		t.WitnessUrl,
		t.MinTotalMissed,
		t.MaxTotalMissed,
		t.StartTime.Format("15:04:05"),
		t.EndTime.Format("15:04:05"),
		t.Duration.String())
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
