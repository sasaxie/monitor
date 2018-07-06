package controllers

import (
	"github.com/astaxie/beego"
)

type BaseController struct {
	beego.Controller
}

func (b *BaseController) Prepare() {
	v := b.GetSession("Validate")
	if v == nil {
		b.Data["json"] = ""
		b.ServeJSON()
	}
}
