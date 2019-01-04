package sender

import (
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/result"
)

type Sender interface {
	Send()
}

func NilSend(results ...result.Result) error {
	logs.Debug("nil sending")
	return nil
}
