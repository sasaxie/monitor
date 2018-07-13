// @APIVersion 1.0.0
// @Title monitor Test API
// @Description monitor is tron-java grpc client
// @TermsOfServiceUrl https://tron.network/
package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/sasaxie/monitor/controllers"
)

func init() {
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))

	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/wsmonitor",
			beego.NSInclude(
				&controllers.WsMonitorController{},
			),
		),
		beego.NSNamespace("/user",
			beego.NSInclude(
				&controllers.UserInfoController{},
			),
		),
		beego.NSNamespace("/node",
			beego.NSInclude(
				&controllers.NodeController{},
			),
		),
		beego.NSNamespace("/program",
			beego.NSInclude(
				&controllers.ProgramController{},
			),
		),
		beego.NSNamespace("/server-group-config",
			beego.NSInclude(
				&controllers.ServerGroupConfigController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
