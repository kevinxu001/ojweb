package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/astaxie/beego"
)

type CommonController struct {
	beego.Controller
}

type Rsp struct {
	Success bool
	Msg     string
}

//create md5 string
func StrToMD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	rs := hex.EncodeToString(h.Sum(nil))
	return rs
}

func (this *CommonController) Prepare() {
	this.Data["SiteName"] = beego.AppConfig.String("site_name")

	path := this.Ctx.Request.RequestURI
	if path == "/" || path == "/reg" || path == "/left" || path == "/right" || path == "/faq" || path == "/check" {
		return
	} else {
		//判断是不是管理员登录
		vadmin := this.GetSession("adminUser")
		if admin_user, ok := vadmin.(string); ok {
			if admin_user != beego.AppConfig.String("admin_user") {
				this.Redirect("/", 302)
				return
			} else {
				admin_realname := beego.AppConfig.String("admin_realname")
				this.Data["IsAdminUser"] = true
				this.Data["Realname"] = admin_realname
				return
			}
		}

		//判断是否是普通用户登录
		v := this.GetSession("currentUser")
		if current_user, ok := v.(string); !ok {
			this.Redirect("/", 302)
			return
		} else {
			this.Data["Realname"] = current_user
			return
		}
	}
}
