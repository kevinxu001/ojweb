package routers

import (
	"github.com/kevinxu001/ojweb/controllers"
	"github.com/astaxie/beego"
)

func init() {
    beego.Router("/", &controllers.MainController{})
}
