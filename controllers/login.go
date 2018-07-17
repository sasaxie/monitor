package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"strings"
)

type UserInfoController struct {
	beego.Controller
}

var (
	userNameValidate = beego.AppConfig.String("username")
	passwordValidate = beego.AppConfig.String("password")
)

// @router /login/:UserName/:Password [get,post]
func (c *UserInfoController) Login() {

	userName := c.GetString(":UserName", "")
	password := c.GetString(":Password", "")

	if strings.EqualFold(userNameValidate, userName) && strings.EqualFold(passwordValidate, password) {
		c.SetSession("Validate", userName)
		c.Data["json"] = "success"
	} else {
		c.Data["json"] = "error"
	}

	c.ServeJSON()
}

// @router /logout [get,post]
func (c *UserInfoController) Logout() {
	//获得id
	//设置返回对象。
	//if c.CruSession == nil {
	//	c.StartSession()
	//}
	//sessionId := c.CruSession.SessionID()
	//logs.Info("==sessionId %s ==", sessionId)
	////设置 SessionDomain 名称。
	//c.DestroySession()
	////设置返回对象。
	//c.Ctx.Redirect(302, "/auth/login")

	c.DelSession("Validate")
	c.DestroySession()
	fmt.Println(c.GetSession("Validate"))
	if c.GetSession("Validate") == nil {
		fmt.Println("返回为nil")
		c.Data["json"] = "success"
	}
	c.ServeJSON()

}
