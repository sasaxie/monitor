package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/sasaxie/monitor/common/config"
)

type NodeInfoController struct {
	beego.Controller
}

// @router /:ip [get]
func (n *NodeInfoController) Get() {
	ip := n.Ctx.Input.Param(":ip")
	address := fmt.Sprintf("%s:%s", ip, config.GRPCDefaultPort)
	fmt.Println(address)
	n.ServeJSON()
}
