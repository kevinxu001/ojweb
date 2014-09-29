package routers

import (
	"github.com/astaxie/beego"
	"github.com/kevinxu001/ojweb/controllers"
)

func init() {
	beego.Router("/", &controllers.IndexController{}, "get:Index")
	beego.Router("/left", &controllers.IndexController{}, "get:Left")
	beego.Router("/right", &controllers.IndexController{}, "get:Right")
	beego.Router("/check", &controllers.IndexController{}, "post:Check")
	beego.Router("/logout", &controllers.IndexController{}, "get:Logout")
	beego.Router("/match", &controllers.IndexController{}, "*:Match")
	beego.Router("/reg", &controllers.IndexController{}, "*:Reg")
	beego.Router("/problem", &controllers.IndexController{}, "get:Problem")
	beego.Router("/submit", &controllers.IndexController{}, "get:Submit")
	beego.Router("/adminproblem", &controllers.IndexController{}, "get:Problem")
	beego.Router("/status", &controllers.IndexController{}, "get:Status")
	beego.Router("/standing", &controllers.IndexController{}, "get:Standing")
	beego.Router("/faq", &controllers.IndexController{}, "get:Faq")
}
