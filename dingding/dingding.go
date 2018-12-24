package dingding

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/sasaxie/monitor/common/config"
	"github.com/sasaxie/monitor/common/http"
)

var DingAlarm *Alarm

func init() {
	DingAlarm = new(Alarm)
	DingAlarm.Url = config.MonitorConfig.Task.Dingding
	beego.Info("init ding alarm, url:", DingAlarm.Url)
}

type Alarm struct {
	Url string
}

func (d *Alarm) Alarm(content []byte) {
	header := make(map[string]string)

	header["Content-Type"] = "application/json"

	body, err := http.Post(content, d.Url, header)

	if err != nil {
		logs.Warn("dingDing alarm post error:", err)
		return
	}

	logs.Info("alarm:", string(body))
}
