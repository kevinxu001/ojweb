package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

type IndexController struct {
	beego.Controller
}

func (this *IndexController) Index() {
	this.TplNames = "index.html"
}

func (this *IndexController) Left() {
	this.TplNames = "left.html"

	//比赛开始时间读取
	var durationSeconds, startTime int64
	if _, err := os.Stat("time.txt"); err == nil {
		ftime, err := os.OpenFile("time.txt", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
		}
		defer ftime.Close()

		fmt.Fscanf(ftime, "%d", &startTime)
	}

	durationSeconds = time.Now().Unix() - startTime

	if startTime > 0 {
		this.Data["IsMatchStart"] = true
	}
	this.Data["DurationSeconds"] = durationSeconds

	//用户登陆状态
	vadmin := this.GetSession("adminUser")
	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.Data["IsLogin"] = true
		this.Data["IsAdmin"] = true

		this.Data["UserName"] = beego.AppConfig.String("admin_realname")
	}

	v := this.GetSession("currentUser")
	if current_user, ok := v.(string); ok {
		this.Data["IsLogin"] = true
		this.Data["IsUser"] = true

		this.Data["UserName"] = current_user
	}

}

func (this *IndexController) Right() {
	this.TplNames = "right.html"

	if _, err := os.Stat("time.txt"); err != nil {
		this.Data["MatchStop"] = true
	}
}

func (this *IndexController) Check() {
	this.TplNames = "check.html"

	username := this.GetString("username")
	//密码未加密
	password := this.GetString("password")
	encryptPassword := StrToMD5(password)

	if username == "" || password == "" {
		this.Data["IsError"] = true
		this.Data["ErrorMsg"] = "用户名或密码为空"
		return
	}

	//判断超级管理员登录
	if username == beego.AppConfig.String("admin_user") {
		if encryptPassword == beego.AppConfig.String("admin_pass") {
			this.SetSession("adminUser", username)
			// this.SetSession("adminRealname", beego.AppConfig.String("admin_realname"))

			this.Data["UserName"] = beego.AppConfig.String("admin_realname")
			return
		}
	}

	//判断普通用户登录
	//'teacher'=8d788385431273d11e8b43bb78f3aa41
	fuser, err := os.OpenFile("user.txt", os.O_RDONLY, 0644)
	if err != nil {
		beego.Error(err)
		this.Data["IsError"] = true
		this.Data["ErrorMsg"] = "无法打开 user.txt 用户数据文件"
		return
	}
	defer fuser.Close()

	//'teacher'=8d788385431273d11e8b43bb78f3aa41
	reguserinfo := regexp.MustCompile(`\'([\p{Han}]|[0-9A-Za-z])*\'=.*\n`)
	var finduserinfo string

	// bf := bufio.NewReader(fuser)
	userdata, err := ioutil.ReadAll(fuser)
	finduserinfo = reguserinfo.FindString(string(userdata))

	uname := regexp.MustCompile(`\'([\p{Han}]|[0-9A-Za-z])*\'`).FindString(finduserinfo)
	uencpass := regexp.MustCompile(`\'=.*`).FindString(finduserinfo)

	if "'"+username+"'" == uname {
		if "'="+encryptPassword == uencpass {
			this.SetSession("currentUser", username)
			// this.SetSession("adminRealname", beego.AppConfig.String("admin_realname"))

			this.Data["UserName"] = username
			return
		}
	}

	this.Data["IsError"] = true
	this.Data["ErrorMsg"] = "用户名或密码错误"

}

func (this *IndexController) Logout() {
	this.TplNames = "logout.html"

	vadmin := this.GetSession("adminUser")
	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.DelSession("adminUser")
		this.DelSession("adminRealname")

		this.Data["UserName"] = admin_user
		return
	}

	v := this.GetSession("currentUser")
	if current_user, ok := v.(string); ok {
		this.DelSession("currentUser")

		this.Data["UserName"] = current_user
		return
	}
}

func (this *IndexController) Match() {
	this.TplNames = "match.html"

	//用户登陆状态
	vadmin := this.GetSession("adminUser")
	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.Data["IsLogin"] = true
		this.Data["IsAdmin"] = true

		// this.Data["UserName"] = admin_user

		//判断time.txt文件是否存在，有就删除，没有就写入当前Unix时间
		if _, err := os.Stat("time.txt"); err == nil {
			if err := os.Remove("time.txt"); err != nil {
				beego.Error(err)
			}
		} else {
			ftime, err := os.OpenFile("time.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				beego.Error(err)
			}
			defer ftime.Close()

			fmt.Fprintln(ftime, time.Now().Unix())
		}
	}

}

func (this *IndexController) Reg() {
	this.TplNames = "reg.html"

	this.Data["AdminUserName"] = beego.AppConfig.String("admin_user")
	beego.Info(this.Ctx.Request.Method)
	if this.Ctx.Request.Method == "GET" {
		this.Data["IsGet"] = true
	} else if this.Ctx.Request.Method == "POST" {
		this.Data["IsPost"] = true
		username := this.GetString("user")
		//密码未加密
		password := this.GetString("pass")
		encryptPassword := StrToMD5(password)

		if username == beego.AppConfig.String("admin_user") {
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "非法用户名"

			return
		}

		execCmd := exec.Command("newuser.exe", username, encryptPassword)
		execOut, err := execCmd.Output()
		if err != nil {
			panic(err)
		}
		this.Data["ExecOut"] = string(execOut)
	}

}

func (this *IndexController) Problem() {
	this.TplNames = "problem.html"

	//比赛开始时间读取
	var startTime int64
	if _, err := os.Stat("time.txt"); err == nil {
		ftime, err := os.OpenFile("time.txt", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
		}
		defer ftime.Close()

		fmt.Fscanf(ftime, "%d", &startTime)
	}

	if startTime > 0 {
		this.Data["IsMatchStart"] = true
	} else {
		return
	}

	vadmin := this.GetSession("adminUser")
	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.Data["IsLogin"] = true
		this.Data["IsAdmin"] = true
		this.Data["UserName"] = admin_user
	}

	v := this.GetSession("currentUser")
	if current_user, ok := v.(string); ok {
		this.Data["IsLogin"] = true
		this.Data["IsUser"] = true
		this.Data["UserName"] = current_user
	}

	probName := this.GetString("probname")
	if probName == "" {
		this.Data["IsProbList"] = true
		//读取题目列表
		if _, err := os.Stat("prob.txt"); err == nil {
			fprob, err := os.OpenFile("prob.txt", os.O_RDONLY, 0644)
			if err != nil {
				beego.Error(err)
			}
			defer fprob.Close()

			var probnum int
			fmt.Fscanf(fprob, "%d", &probnum)

			var hasProbset bool
			var fprobset *os.File
			if _, err := os.Stat("probset.txt"); err == nil {
				fprobset, err := os.OpenFile("probset.txt", os.O_RDONLY, 0644)
				if err != nil {
					beego.Error(err)
				}
				defer fprobset.Close()

				hasProbset = true
			}

			type ProbInfo struct {
				Id             int
				Class          int
				ProbName       string
				AcTotal        int
				SubTotal       int
				CorrectPercent float32
			}

			probInfos := make([]*ProbInfo, 0, probnum)

			id := 0
			for {
				probinfo := new(ProbInfo)

				n, _ := fmt.Fscanf(fprob, "%s", &probinfo.ProbName)
				if n == 0 {
					break
				}

				id++
				probinfo.Id = id

				if hasProbset {
					n, _ = fmt.Fscanf(fprobset, "%d %d", &probinfo.AcTotal, &probinfo.SubTotal)
				}

				if probinfo.SubTotal > 0 {
					probinfo.CorrectPercent = float32(probinfo.AcTotal * 100.0 / probinfo.SubTotal)
				}

				probInfos = append(probInfos, probinfo)
			}

			this.Data["ProbInfos"] = probInfos
		}
	} else {
		this.Data["ProbName"] = probName

		if _, err := os.Stat("prob/" + probName + "/prob.html"); err != nil {
			beego.Error(err)
			this.Data["ProbHtml"] = "找不到题目描述文件"
			return
		} else {
			fprobhtml, err := os.OpenFile("prob/"+probName+"/prob.html", os.O_RDONLY, 0644)
			if err != nil {
				beego.Error(err)
				this.Data["ProbHtml"] = "找不到题目描述文件"
				return
			}
			defer fprobhtml.Close()

			var probhtml []byte
			probhtml, _ = ioutil.ReadAll(fprobhtml)
			this.Data["ProbHtml"] = string(probhtml)
		}
	}

}

func (this *IndexController) Submit() {
	this.TplNames = "submit.html"

	//比赛开始时间读取
	var startTime int64
	if _, err := os.Stat("time.txt"); err == nil {
		ftime, err := os.OpenFile("time.txt", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
		}
		defer ftime.Close()

		fmt.Fscanf(ftime, "%d", &startTime)
	}

	if startTime > 0 {
		this.Data["IsMatchStart"] = true
	} else {
		return
	}

	vadmin := this.GetSession("adminUser")
	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.Data["IsLogin"] = true
		this.Data["IsAdmin"] = true
		this.Data["UserName"] = admin_user
	}

	v := this.GetSession("currentUser")
	if current_user, ok := v.(string); ok {
		this.Data["IsLogin"] = true
		this.Data["IsUser"] = true
		this.Data["UserName"] = current_user
	}

	probName := this.GetString("probname")
	if probName == "" {
		this.Data["IsProbList"] = true
		//读取题目列表
		if _, err := os.Stat("prob.txt"); err == nil {
			fprob, err := os.OpenFile("prob.txt", os.O_RDONLY, 0644)
			if err != nil {
				beego.Error(err)
				this.Data["IsError"] = true
				this.Data["ErrorMsg"] = "找不到题目！"
			}
			defer fprob.Close()

			var probnum int
			fmt.Fscanf(fprob, "%d", &probnum)

			type ProbInfo struct {
				Id       int
				ProbName string
			}

			probInfos := make([]*ProbInfo, 0, probnum)

			id := 0
			for {
				probinfo := new(ProbInfo)

				n, _ := fmt.Fscanf(fprob, "%s", &probinfo.ProbName)
				if n == 0 {
					break
				}

				id++
				probinfo.Id = id

				probInfos = append(probInfos, probinfo)
			}

			this.Data["ProbInfos"] = probInfos
		}
	} else if this.Ctx.Request.Method == "POST" {
		this.Data["ProbName"] = probName

		if _, err := os.Stat("prob/" + probName + "/prob.html"); err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "找不到题目！"
			return
		}

		lang := this.GetString("lang")

		// file, fileHeader, err := this.GetFile("program")
		file, _, err := this.GetFile("program")
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "没有上传文件！"
			return
		}
		defer file.Close()

		//读入最后更新ID
		ftail, err := os.OpenFile("tail.txt", os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "找不到 tail.txt 文件！"
			return
		}
		defer ftail.Close()

		var tailid int
		fmt.Fscanf(ftail, "%d", &tailid)
		tailid++
		fmt.Fprintf(ftail, "%d", tailid)

		// curdir, err := os.Getwd()

		tofilepath := "submit/" + strconv.Itoa(tailid) + "/"
		err = os.MkdirAll(tofilepath, 0755)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "创建 " + tofilepath + " 失败！"
			return
		}

		ftime, err := os.OpenFile("time.txt", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "找不到 time.txt 文件！"
			return
		}
		defer ftime.Close()

		var startTime int
		fmt.Fscanf(ftime, "%d", &startTime)

		ntime := (time.Now().Unix() - int64(startTime)) / 60

		fsubtime, err := os.OpenFile(tofilepath+"time.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "无法创建 " + tofilepath + "time.txt 文件！"
			return
		}
		defer fsubtime.Close()

		fmt.Fprintf(fsubtime, "%d", ntime)

		fsubresult, err := os.OpenFile(tofilepath+"result.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "无法创建 " + tofilepath + "result.txt 文件！"
			return
		}
		defer fsubresult.Close()

		fmt.Fprintf(fsubresult, "%s", "Waiting")

		fsubuser, err := os.OpenFile(tofilepath+"user.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "无法创建 " + tofilepath + "user.txt 文件！"
			return
		}
		defer fsubuser.Close()

		fmt.Fprintf(fsubuser, "%s", this.Data["UserName"])

		fsubprob, err := os.OpenFile(tofilepath+"prob.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "无法创建 " + tofilepath + "prob.txt 文件！"
			return
		}
		defer fsubprob.Close()

		fmt.Fprintf(fsubprob, "%s", probName)

		fsublang, err := os.OpenFile(tofilepath+"lang.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "无法创建 " + tofilepath + "lang.txt 文件！"
			return
		}
		defer fsublang.Close()

		var tofile string
		if lang == "1" {
			tofile = tofilepath + "prog.cpp"
			fmt.Fprintf(fsublang, "%s", "C++")
		} else if lang == "2" {
			tofile = tofilepath + "prog.pas"
			fmt.Fprintf(fsublang, "%s", "PASCAL")
		}

		f, err := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "创建 " + tofile + " 失败！"
			return
		}
		defer f.Close()

		io.Copy(f, file)

	}

}

func (this *IndexController) AdminProblem() {
	this.TplNames = "adminproblem.html"

	vadmin := this.GetSession("adminUser")
	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.Data["IsLogin"] = true
		this.Data["IsAdmin"] = true
		this.Data["UserName"] = admin_user

		return
	} else {
		this.Data["IsError"] = true
		this.Data["ErrorMsg"] = "请使用管理员帐号登录"

		return
	}
}

func (this *IndexController) Status() {
	this.TplNames = "status.html"

	//比赛开始时间读取
	var startTime int64
	if _, err := os.Stat("time.txt"); err == nil {
		ftime, err := os.OpenFile("time.txt", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
		}
		defer ftime.Close()

		fmt.Fscanf(ftime, "%d", &startTime)
	}

	if startTime > 0 {
		this.Data["IsMatchStart"] = true
	} else {
		return
	}

	vadmin := this.GetSession("adminUser")
	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.Data["IsLogin"] = true
		this.Data["IsAdmin"] = true
		this.Data["UserName"] = admin_user
	}

	v := this.GetSession("currentUser")
	if current_user, ok := v.(string); ok {
		this.Data["IsLogin"] = true
		this.Data["IsUser"] = true
		this.Data["UserName"] = current_user
	}

	if _, err := os.Stat("status.html"); err != nil {
		beego.Error(err)
		this.Data["StatusHtml"] = "找不到 status.html 文件"
		return
	} else {
		fstatushtml, err := os.OpenFile("status.html", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["StatusHtml"] = "找不到 status.html 文件"
			return
		}
		defer fstatushtml.Close()

		var statushtml []byte
		statushtml, _ = ioutil.ReadAll(fstatushtml)
		this.Data["StatusHtml"] = string(statushtml)
	}
}

func (this *IndexController) Standing() {
	this.TplNames = "standing.html"

	//比赛开始时间读取
	var startTime int64
	if _, err := os.Stat("time.txt"); err == nil {
		ftime, err := os.OpenFile("time.txt", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
		}
		defer ftime.Close()

		fmt.Fscanf(ftime, "%d", &startTime)
	}

	if startTime > 0 {
		this.Data["IsMatchStart"] = true
	} else {
		return
	}

	vadmin := this.GetSession("adminUser")
	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.Data["IsLogin"] = true
		this.Data["IsAdmin"] = true
		this.Data["UserName"] = admin_user
	}

	v := this.GetSession("currentUser")
	if current_user, ok := v.(string); ok {
		this.Data["IsLogin"] = true
		this.Data["IsUser"] = true
		this.Data["UserName"] = current_user
	}

	if _, err := os.Stat("rank.html"); err != nil {
		beego.Error(err)
		this.Data["RankHtml"] = "找不到 rank.html 文件"
		return
	} else {
		frankhtml, err := os.OpenFile("rank.html", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["RankHtml"] = "找不到 rank.html 文件"
			return
		}
		defer frankhtml.Close()

		var rankhtml []byte
		rankhtml, _ = ioutil.ReadAll(frankhtml)
		this.Data["RankHtml"] = string(rankhtml)
	}

}

func (this *IndexController) Faq() {
	this.TplNames = "faq.html"

}
