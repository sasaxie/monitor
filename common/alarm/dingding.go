package alarm

import (
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/common/mhttp"
	"log"
)

var DingAlarm *DingDingAlarm

func init() {
	DingAlarm = new(DingDingAlarm)
	DingAlarm.Url = beego.AppConfig.String("dingdingURl")
	beego.Info("Init Ding alarm, url:", DingAlarm.Url)
}

type DingDingAlarm struct {
	Url string
}

func (d *DingDingAlarm) Alarm(content []byte) {
	header := make(map[string]string)

	header["Content-Type"] = "application/json"

	body, err := mhttp.Post(content, d.Url, header)

	if err != nil {
		beego.Warning("DingDing alarm post error:", err.Error())
		return
	}

	log.Println(string(body))
}
