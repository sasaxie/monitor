package routers

import (
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/controllers"
)

func init() {
	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/node-info",
			beego.NSInclude(
				&controllers.NodeInfoController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
