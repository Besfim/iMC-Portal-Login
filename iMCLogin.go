package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"net/http"
	"strings"
	"io"
	"net/url"
	"encoding/base64"
	"time"
	"strconv"
	"flag"
	"os"
)

var debug = flag.Bool("d",false,"开启 DEBUG 日志打印")

var c *colly.Collector
// 学号
var userNum = flag.String("u","null","学生学号")
// 密码
var userPw = flag.String("p","null","校园网密码")
// 起始联网时间戳
var startTime string
// 心跳包发送间隔,单位是秒
const heartBeatInterval time.Duration = 60 * 5

func main() {
	initImcGDPU()
	debugLog("userName:" + *userNum + ", password: " + *userPw)

	cookie,pl := getCookieAndPL()
	if login(cookie, pl) {
		if isConnect() {
			log("联网成功, 输入 exit 可下线并退出程序")
			go stayConnect(cookie, pl)
			var userInput string
			for ;; {
				fmt.Scanln(&userInput)
				if userInput == "exit" {
					if logout(cookie,pl,startTime) {
						log("下线成功")
					}else {
						debugLog("下线失败, 使用强制下线")
						if ensureLogout() {
							log("下线成功")
						}else {
							log("下线失败")
						}
					}
					log("程序已退出")
					break
				}
			}
		}else {
			log("联网失败, 请检查网络连接")
		}
	}
}

//初始化
func initImcGDPU() {
	flag.Parse()
	if *userNum == "null" || *userPw == "null" {
		fmt.Println("参数为空或参数错误, 请使用 -h 查看使用帮助")
		os.Exit(0)
	}else {
		c = colly.NewCollector()
		colly.Async(false)
		// TODO 作者标记
		// TODO 获取新版本
		// TODO 强制下线的 flag
	}
}

//获取 cookie
func getCookieAndPL() (string, string) {
	var cookie string
	var pl string
	header := http.Header{
		"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7"},
	}
	c.OnResponse(func(response *colly.Response) {
		cookie = response.Headers.Get("Set-Cookie")
		if cookie == "" {
			debugLog("登录页面返回 cookie : "+cookie)
			log("登录页面访问失败, 程序退出")
			os.Exit(0)
		}else {
			pl = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
		}
	})
	c.Request("GET", "http://10.50.15.9/portal/templatePage/20170110154814101/login_custom.jsp?userip=", nil, nil, header)
	return cookie, pl
}

//登录
func login(cookie string, pl string) bool {

	successFlag := false

	header := http.Header{
		"Cookie":          []string{cookie},
		"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7"},
	}
	formDate := map[string]string{
		"userName":            *userNum,
		"userPwd":             base64.StdEncoding.EncodeToString([]byte(*userPw)),
		"userDynamicPwd":      "",
		"userDynamicPwdd":     "",
		"serviceTypeHIDE":     "",
		"serviceType":         "",
		"userurl":             "",
		"userip":              "",
		"basip":               "",
		"language":            "Chinese",
		"usermac":             "null",
		"wlannasid":           "",
		"wlanssid":            "",
		"entrance":            "null",
		"loginVerifyCode":     "",
		"userDynamicPwddd":    "",
		"customPageId":        "105",
		"pwdMode":             "0",
		"portalProxyIP":       "10.50.15.9",
		"portalProxyPort":     "50200",
		"dcPwdNeedEncrypt":    "1",
		"assignIpType":        "0",
		"appRootUrl":          "http://10.50.15.9/portal/",
		"manualUrl":           "",
		"manualUrlEncryptKey": "",
	}
	c.OnResponse(func(response *colly.Response) {
		info := string(response.Body)
		decodeTemp, _ := base64.StdEncoding.DecodeString(info)
		respJson, _ := url.QueryUnescape(string(decodeTemp))
		debugLog("登录返回 : " + respJson)
		startTime = strconv.FormatInt(time.Now().UnixNano(),10)[0:13]
		successFlag = true
		// TODO 对返回的判断
		// TODO 如果返回登录服务器错误的话证明之前有在线,现在已经下线,所以应该再登录一次
	})
	c.Request(
		"POST",
		"http://10.50.15.9/portal/templatePage/20170110154814101/login_custom.jsp?userip=",
		createFormReader(formDate),
		nil,
		header,
	)
	return successFlag
}

func heartBeat(cookie string, pl string) {
	c.OnResponse(func(response *colly.Response) {
		debugLog("成功发送了一个心跳包")
	})
	header := http.Header{
		"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7"},
		"Cookie":		   []string{cookie},
		"Referer":         []string{"http://10.50.15.9/portal/page/online.jsp?st=2&pl=" + pl + "&custompath=templatePage/20170110154814101/&uamInitCustom=0&customCfg=MTA1&uamInitLogo=H3C&userName=null&userPwd=null&loginType=3&innerStr=null&outerStr=null&v_is_selfLogin=0"},
	}
	c.Request("GET", "http://10.50.15.9/portal/page/online_heartBeat.jsp?pl="+pl+"&custompath=templatePage/20170110154814101/&uamInitCustom=0&uamInitLogo=H3C", nil, nil, header)
}

func logout(cookie string, pl string, startTime string) bool {
	resFlag := false
	c.OnResponse(func(response *colly.Response) {
		debugLog("成功发送了退出请求")
		info := string(response.Body)
		decodeTemp, _ := base64.StdEncoding.DecodeString(info)
		respJson, _ := url.QueryUnescape(string(decodeTemp))
		// 如果可以可以 GET 这个地址并且得到返回的话, 就证明之前是在线的, 并且现在退出了
		// 返回的 json 是 {"errorNumber":"1"}
		debugLog(respJson)
		if strings.Contains(respJson,"{\"errorNumber\":\"1\"}") {
			debugLog("退出成功")
			resFlag = true
		}
	})
	header := http.Header{
		"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7"},
		"Cookie":		   []string{"hello1="+*userNum+"; hello2=false;"+cookie},
		"Referer":         []string{"http://10.50.15.9/portal/page/online.jsp?st=2&pl=" + pl + "&custompath=templatePage/20170110154814101/&uamInitCustom=0&customCfg=MTA1&uamInitLogo=H3C&userName=null&userPwd=null&loginType=3&innerStr=null&outerStr=null&v_is_selfLogin=0"},
	}
	c.Request("GET", "http://10.50.15.9/portal/pws?t=lo&language=Chinese&userip=&basip=&_="+startTime, nil, nil, header)
	return resFlag
}

//这个方法是用来确保程序已经退出的
func ensureLogout() bool {
	// TODO 登录一次, 如果登录显示 "服务器错误" 证明下线
	// TODO 如果能够成功登录, 那么就证明之前已经下线, 所以需要直接调用 logout 方法下线
	return false
}

func isConnect() bool {
	connectFlag := false
	c.OnResponse(func(response *colly.Response) {
		if strings.Contains(string(response.Body), "百度一下，你就知道") {
			connectFlag = true
		}
	})
	header := http.Header{
		"User-Agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
	}
	c.Request("GET", "https://www.baidu.com/", nil, nil, header)
	if connectFlag {
		debugLog("网络联通")
	} else {
		debugLog("网络不联通")
	}
	return connectFlag
}

func createFormReader(data map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}
	return strings.NewReader(form.Encode())
}

var reConnectTime = 0;
func stayConnect(cookie string, pl string)  {
	for ;; {
		if reConnectTime > 10 {
			log("连接失败次数过多, 请确认是否联网")
		}
		heartBeat(cookie, pl)
		if !isConnect() {
			log("发送心跳包后联网失败, 尝试重新连接")
			reConnectTime++
			cookie,pl := getCookieAndPL()
			login(cookie,pl)
			if isConnect() {
				log("联网成功")
			}else {
				continue
			}
		}else {
			reConnectTime = 0
		}
		time.Sleep(time.Second * heartBeatInterval)
	}
}


// 打印 debug 日志
func debugLog(msg string) {
	if *debug {
		fmt.Println("[LOG] [DEBUG]   ["+getTime(time.Now())+"]:", msg)
	}
}

func log(msg string) {
	fmt.Println("[LOG] [RUNNING] ["+getTime(time.Now())+"]:",msg)
}

func getTime(time time.Time) string {
	return time.String()[5:19]
}