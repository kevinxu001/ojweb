package controllers

import (
	"bufio"
	"fmt"
	"github.com/astaxie/beego"
	"strings"
	// iconv "github.com/djimenez/iconv-go"
	"code.google.com/p/mahonia"
	"io"
	"io/ioutil"
	"os"
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
	reguserinfo := regexp.MustCompile(`\'` + username + `\'=.*\n`)
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
	// beego.Info(this.Ctx.Request.Method)
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

		if len(username) < 4 {
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "用户名太短。至少4个英文字符或2个中文字符。"
			return
		}

		if ok, _ := regexp.Match(`^([\p{Han}]|[0-9A-Za-z])*$`, []byte(username)); !ok {
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "用户名\"" + username + "\"含有非法字符。"
			return
		}

		if len(password) < 6 {
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "密码太短。至少6个英文字符。"
			return
		}

		// curdir, _ := os.Getwd()

		f, err := os.OpenFile("user.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "创建 user.txt 失败！"
			return
		}
		defer f.Close()

		//'teacher'=8d788385431273d11e8b43bb78f3aa41
		bf := bufio.NewReader(f)
		reguserinfo := regexp.MustCompile(`\'([\p{Han}]|[0-9A-Za-z])*\'`)
		var finduser string
		for {
			//buf,err := r.ReadBytes('\n')
			buf, err := bf.ReadString('\n')
			if err == io.EOF {
				break
			}

			finduser = reguserinfo.FindString(buf)

			if finduser == `'`+username+`'` {
				this.Data["IsError"] = true
				this.Data["ErrorMsg"] = "用户名\"" + username + "\"已经存在，请更换用户名。"
				return
			}
		}

		fmt.Fprintf(f, "'%s'=%s\n", username, encryptPassword)

		this.Data["UserName"] = username
	}

}

func (this *IndexController) Problem() {
	this.TplNames = "problem.html"

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

			var buf string
			n, _ := fmt.Fscanf(fprob, "%s\r", &buf)
			isWin := n > 0
			if isWin {
				fprob.Seek(0, os.SEEK_SET)
			}

			var probnum int
			bfprob := bufio.NewReader(fprob)
			ret, _ := bfprob.ReadString('\n')
			if isWin {
				if n := strings.LastIndex(ret, "\r\n"); n >= 0 {
					ret = ret[:n]
				}
			}
			probnum, _ = strconv.Atoi(ret)

			var hasProbset bool
			var fprobset *os.File
			if _, err := os.Stat("probset.txt"); err == nil {
				fprobset, err = os.OpenFile("probset.txt", os.O_RDONLY, 0644)
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

			var pname string
			id := 0
			for {
				probinfo := new(ProbInfo)

				//兼容WIN下的换行符号读入
				// fmt.Fscanf(fprob, "%d", &buf)
				// if buf != 0 {
				// 	fprob.Seek(-1, os.SEEK_CUR)
				// }

				pname, err = bfprob.ReadString('\n')
				if err == io.EOF && len(pname) == 0 {
					break
				}

				//转换WIN下gb18030 to utf-8
				if isWin {
					// pname=strings.TrimSpace(pname)
					if n := strings.LastIndex(pname, "\r\n"); n >= 0 {
						pname = pname[:n]
					}
					dec := mahonia.NewDecoder("gb18030")
					probinfo.ProbName = dec.ConvertString(pname)
				}

				id++
				probinfo.Id = id

				if hasProbset {
					_, err = fmt.Fscanf(fprobset, "%d %d\n", &probinfo.AcTotal, &probinfo.SubTotal)
					// beego.Info(probinfo.AcTotal, probinfo.SubTotal)
				}

				if probinfo.SubTotal > 0 {
					probinfo.CorrectPercent = float32(probinfo.AcTotal * 100.0 / probinfo.SubTotal)
				}

				probInfos = append(probInfos, probinfo)
			}
			// converter.Close()

			this.Data["ProbInfos"] = probInfos
		}
	} else {
		this.Data["ProbName"] = probName

		fprob, err := os.OpenFile("prob.txt", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
		}
		defer fprob.Close()
		var buf string
		n, _ := fmt.Fscanf(fprob, "%s\r", &buf)
		isWin := n > 0

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
			//转换WIN下gb18030 to utf-8
			if isWin {
				dec := mahonia.NewDecoder("gb18030")
				this.Data["ProbHtml"] = dec.ConvertString(string(probhtml))
			} else {
				this.Data["ProbHtml"] = string(probhtml)
			}

		}
	}

}

func (this *IndexController) Submit() {
	this.TplNames = "submit.html"

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
			var buf string
			n, _ := fmt.Fscanf(fprob, "%s\r", &buf)
			isWin := n > 0
			if isWin {
				fprob.Seek(0, os.SEEK_SET)
			}

			bfprob := bufio.NewReader(fprob)
			ret, _ := bfprob.ReadString('\n')
			probnum, _ := strconv.Atoi(ret)

			type ProbInfo struct {
				Id       int
				ProbName string
			}

			probInfos := make([]*ProbInfo, 0, probnum)

			var pname string
			id := 0
			for {
				probinfo := new(ProbInfo)

				pname, err = bfprob.ReadString('\n')
				if err == io.EOF && len(pname) == 0 {
					break
				}

				//转换WIN下gb18030 to utf-8
				if isWin {
					if n := strings.LastIndex(pname, "\r\n"); n >= 0 {
						pname = pname[:n]
					}
					dec := mahonia.NewDecoder("gb18030")
					probinfo.ProbName = dec.ConvertString(pname)
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
		ftailr, err := os.OpenFile("tail.txt", os.O_RDONLY, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "找不到 tail.txt 文件！"
			return
		}
		defer ftailr.Close()

		var tailid int
		fmt.Fscanf(ftailr, "%d", &tailid)

		ftailw, err := os.OpenFile("tail.txt", os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "找不到 tail.txt 文件！"
			return
		}
		defer ftailw.Close()
		tailid++
		fmt.Fprintf(ftailw, "%d", tailid)

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

		fsubip, err := os.OpenFile(tofilepath+"ip.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			beego.Error(err)
			this.Data["IsError"] = true
			this.Data["ErrorMsg"] = "无法创建 " + tofilepath + "ip.txt 文件！"
			return
		}
		defer fsubip.Close()

		fmt.Fprintf(fsubip, "%s", this.Ctx.Input.IP())

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

		CopyFile(tofilepath+"test.bat", "prob/"+probName+"/test.bat")
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

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

//从指定的文件名中读出string
func readFile(filename string) string {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var str string
	bf := bufio.NewReader(f)
	str, _ = bf.ReadString('\n')
	// fmt.Fscanf(f, "%s\n", &str)

	return str
}

func (this *IndexController) ShowProg() {
	this.TplNames = "showprog.html"

	vadmin := this.GetSession("adminUser")

	if admin_user, ok := vadmin.(string); ok && admin_user == beego.AppConfig.String("admin_user") {
		this.Data["IsLogin"] = true
		this.Data["IsAdmin"] = true
		this.Data["UserName"] = admin_user
	} else {
		this.Data["IsError"] = true
		this.Data["ErrorMsg"] = "你不是老师,没有权限查看!"
		return
	}

	progid := this.GetString("id")

	progfilepath := "submit/" + progid + "/"

	lang := readFile(progfilepath + "lang.txt")

	var prog string
	if lang[:3] == "PAS" {
		prog = progfilepath + "prog.pas"
	} else if lang[:3] == "C++" {
		prog = progfilepath + "prog.cpp"
	}

	// beego.Info(prog)

	proguser := readFile(progfilepath + "user.txt")

	progname := readFile(progfilepath + "prob.txt")

	progip := readFile(progfilepath + "ip.txt")

	fprog, err := os.OpenFile(prog, os.O_RDONLY, 0644)
	if err != nil {
		this.Data["IsError"] = true
		this.Data["ErrorMsg"] = "程序文件不存在!"
		return
	}
	defer fprog.Close()

	var progcontent []byte
	progcontent, _ = ioutil.ReadAll(fprog)

	this.Data["ProgUser"] = proguser
	this.Data["ProgName"] = progname
	this.Data["ProgIp"] = progip
	this.Data["ProgLang"] = lang
	this.Data["ProgId"] = progid
	this.Data["ProgContent"] = string(progcontent)
}
