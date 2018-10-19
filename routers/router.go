package routers

import (
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/controllers"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/monitor-info",
			beego.NSInclude(
				&controllers.MonitorInfoController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
