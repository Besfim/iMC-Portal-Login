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
	"bufio"
)

var debug = flag.Bool("d",false,"开启 DEBUG 日志打印")
// 强制退出
var ensureLogoutFlag = flag.Bool("o",false,"强制退出特定帐号, 也要求正确的用户名和密码")
// 学号
var userNum = flag.String("u","null","学生学号")
// 密码
var userPw = flag.String("p","null","校园网密码")
// 起始联网时间戳
var startTime string
// 心跳包发送间隔,单位是秒
const heartBeatInterval time.Duration = 30

func main() {

	initImcGDPU()
	debugLog("userName:" + *userNum + ", password: " + *userPw)

	cookie,pl := getCookieAndPL()
	if login(cookie, pl) {
		if isConnect() {
			go stayConnect(cookie, pl)
			log("联网成功, 输入 exit 可下线并退出程序")
			go inputExit(cookie, pl)
			select {}
		}else {
			log("联网失败, 请检查网络连接")
		}
	}
}

func inputExit(cookie string, pl string ){
	for {
		inputReader := bufio.NewReader(os.Stdin)
		input, err := inputReader.ReadString('\n')
		if err != nil {
			debugLog("获取用户输入失败, : " + err.Error())
			debugLog("请停止程序之后使用 -o 强制退出, 确保帐号登出")
			break
		}
		if strings.Contains(input,"exit") { // Windows, on Linux it is "S\n"
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
			os.Exit(0)
			break
		}else {
			fmt.Println("输入 exit 可下线并退出程序")
		}
	}
}

//初始化
func initImcGDPU() {
	flag.Parse()
	if *ensureLogoutFlag {
		if ensureLogout() {
			log("退出成功")
		}else {
			log("退出失败, 请使用 -d 开启 debug 模式查看日志")
		}
		os.Exit(0)
	}
	if *userNum == "null" || *userPw == "null" {
		fmt.Println("用户名或密码为空, 或参数错误, 请使用 -h 查看使用帮助")
		os.Exit(0)
	}else {
		colly.Async(false)
		// TODO 作者标记
		// TODO 获取新版本
		// TODO 强制下线的 flag
	}
}

//获取 cookie
func getCookieAndPL() (string, string) {
	c := colly.NewCollector()
	var cookie string
	var pl string
	header := http.Header{
		"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7"},
	}
	c.OnResponse(func(response *colly.Response) {
		cookie = response.Headers.Get("Set-Cookie")
	})
	c.Request("GET", "http://10.50.15.9/portal/templatePage/20170110154814101/login_custom.jsp?userip=", nil, nil, header)
	if cookie == "" {
		debugLog("登录页面返回 cookie : "+cookie)
		log("登录页面访问失败, 程序退出")
		os.Exit(0)
	}else {
		pl = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	}
	return cookie, pl
}

//登录
func login(cookie string, pl string) bool {
	c := colly.NewCollector()
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
		debugLog("登录返回info : " + info)
		decodeTemp := decodeRespInfo(info)
		respJson, _ := url.QueryUnescape(decodeTemp)
		debugLog("登录返回 : " + respJson)
		startTime = strconv.FormatInt(time.Now().UnixNano(),10)[0:13]
		// 登录成功 {"errorNumber":"1","heartBeatCyc":900000,"heartBeatTimeoutMaxTime":2,"userDevPort":"Servers-Aggregation-SW-S5560X-54C-EI-vlan-01-4055@vlan","userStatus":99,"serialNo":4686,"ifNeedModifyPwd":false,"browserUrl":"","clientPrivateIp":"","userurl":"","usermac":null,"nasIp":"","clientLanguage":"Chinese","ifTryUsePopupWindow":true,"triggerRedirectUrl":"","portalLink":"JTdCJTIyZXJyb3JOdW1iZXIlMjIlM0ElMjIxJTIyJTJDJTIyaGVhcnRCZWF0Q3ljJTIyJTNBOTAwMDAwJTJDJTIyaGVhcnRCZWF0VGltZW91dE1heFRpbWUlMjIlM0EyJTJDJTIydXNlckRldlBvcnQlMjIlM0ElMjJTZXJ2ZXJzLUFnZ3JlZ2F0aW9uLVNXLVM1NTYwWC01NEMtRUktdmxhbi0wMS00MDU1JTQwdmxhbiUyMiUyQyUyMnVzZXJTdGF0dXMlMjIlM0E5OSUyQyUyMnNlcmlhbE5vJTIyJTNBNDY4NiUyQyUyMmlmTmVlZE1vZGlmeVB3ZCUyMiUzQWZhbHNlJTJDJTIyYnJvd3NlclVybCUyMiUzQSUyMiUyMiUyQyUyMmNsaWVudFByaXZhdGVJcCUyMiUzQSUyMiUyMiUyQyUyMnVzZXJ1cmwlMjIlM0ElMjIlMjIlMkMlMjJ1c2VybWFjJTIyJTNBbnVsbCUyQyUyMm5hc0lwJTIyJTNBJTIyJTIyJTJDJTIyY2xpZW50TGFuZ3VhZ2UlMjIlM0ElMjJDaGluZXNlJTIyJTJDJTIyaWZUcnlVc2VQb3B1cFdpbmRvdyUyMiUzQXRydWUlMkMlMjJ0cmlnZ2VyUmVkaXJlY3RVcmwlMjIlM0ElMjIlMjIlN0Q"}
		if strings.Contains(respJson, "errorNumber") && strings.Contains(respJson, "heartBeatTimeoutMaxTime") {
			successFlag = true
		}else if strings.Contains(respJson,"E63032:密码错误") {
			// 密码错误 	{"portServIncludeFailedCode":"63032","portServIncludeFailedReason":"E63032:密码错误，您还可以重试8次。","e_c":"portServIncludeFailedCode","e_d":"portServIncludeFailedReason","errorNumber":"7"}
			log("用户密码错误")
			successFlag = false
		}else if strings.Contains(respJson,"E63018:用户不存在或者用户没有申请该服务") {
			// 用户名不存在 {"portServIncludeFailedCode":"63018","portServIncludeFailedReason":"E63018:用户不存在或者用户没有申请该服务。","e_c":"portServIncludeFailedCode","e_d":"portServIncludeFailedReason","errorNumber":"7"}
			log("用户已经登录, 尝试强制下线后重新登录")
			successFlag = false
		}else if strings.Contains(respJson,"设备拒绝请求") {
			// 已经登录		{"portServErrorCode":"1","portServErrorCodeDesc":"设备拒绝请求","e_c":"portServErrorCode","e_d":"portServErrorCodeDesc","errorNumber":"7"}
			log("用户已经登录, 尝试强制下线后重新登录")
			successFlag = login(cookie,pl)
		}else {
			//
			log("登录失败, 错误信息如下: "+respJson)
			successFlag = false
		}
	})
	c.Request(
		"POST",
		"http://10.50.15.9/portal/pws?t=li",
		createFormReader(formDate),
		nil,
		header,
	)
	return successFlag
}

func heartBeat(cookie string, pl string) {
	c := colly.NewCollector()
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
	c := colly.NewCollector()
	resFlag := false
	c.OnResponse(func(response *colly.Response) {
		debugLog("成功发送了退出请求")
		info := string(response.Body)
		decodeTemp := decodeRespInfo(info)
		respJson, _ := url.QueryUnescape(decodeTemp)
		// 如果可以可以 GET 这个地址并且得到返回的话, 就证明之前是在线的, 并且现在退出了
		// 返回的 json 是 {"errorNumber":"1"}
		debugLog(respJson)
		if strings.Contains(respJson,"errorNumber") {
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
	cookie,pl := getCookieAndPL()
	c := colly.NewCollector()
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
		debugLog("登录返回info : " + info)
		decodeTemp := decodeRespInfo(info)
		respJson, _ := url.QueryUnescape(decodeTemp)
		debugLog("登录返回 : " + respJson)
		startTime = strconv.FormatInt(time.Now().UnixNano(),10)[0:13]
		// 登录成功 {"errorNumber":"1","heartBeatCyc":900000,"heartBeatTimeoutMaxTime":2,"userDevPort":"Servers-Aggregation-SW-S5560X-54C-EI-vlan-01-4055@vlan","userStatus":99,"serialNo":4686,"ifNeedModifyPwd":false,"browserUrl":"","clientPrivateIp":"","userurl":"","usermac":null,"nasIp":"","clientLanguage":"Chinese","ifTryUsePopupWindow":true,"triggerRedirectUrl":"","portalLink":"JTdCJTIyZXJyb3JOdW1iZXIlMjIlM0ElMjIxJTIyJTJDJTIyaGVhcnRCZWF0Q3ljJTIyJTNBOTAwMDAwJTJDJTIyaGVhcnRCZWF0VGltZW91dE1heFRpbWUlMjIlM0EyJTJDJTIydXNlckRldlBvcnQlMjIlM0ElMjJTZXJ2ZXJzLUFnZ3JlZ2F0aW9uLVNXLVM1NTYwWC01NEMtRUktdmxhbi0wMS00MDU1JTQwdmxhbiUyMiUyQyUyMnVzZXJTdGF0dXMlMjIlM0E5OSUyQyUyMnNlcmlhbE5vJTIyJTNBNDY4NiUyQyUyMmlmTmVlZE1vZGlmeVB3ZCUyMiUzQWZhbHNlJTJDJTIyYnJvd3NlclVybCUyMiUzQSUyMiUyMiUyQyUyMmNsaWVudFByaXZhdGVJcCUyMiUzQSUyMiUyMiUyQyUyMnVzZXJ1cmwlMjIlM0ElMjIlMjIlMkMlMjJ1c2VybWFjJTIyJTNBbnVsbCUyQyUyMm5hc0lwJTIyJTNBJTIyJTIyJTJDJTIyY2xpZW50TGFuZ3VhZ2UlMjIlM0ElMjJDaGluZXNlJTIyJTJDJTIyaWZUcnlVc2VQb3B1cFdpbmRvdyUyMiUzQXRydWUlMkMlMjJ0cmlnZ2VyUmVkaXJlY3RVcmwlMjIlM0ElMjIlMjIlN0Q"}
		if strings.Contains(respJson,"errorNumber") && strings.Contains(respJson, "heartBeatTimeoutMaxTime") {
			log("尝试登录成功")
			// 似乎这里的 logout 一直是没法使用的
			tempFlag := logout(cookie,pl,getTime(time.Now()))
			if tempFlag {
				debugLog("下线成功")
				successFlag = true
			}else {
				debugLog("正常下线失败, 尝试重新强制下线")
				//如果这个时候下线不了的话就来多一次登录
				successFlag = ensureLogout()
			}
		}else if strings.Contains(respJson,"设备拒绝请求") {
			debugLog("强制下线成功")
			successFlag = true
		}else {
			//
			log("强制下线失败, 尝试登录后错误信息如下: "+respJson)
			successFlag = false
		}
	})
	c.Request(
		"POST",
		"http://10.50.15.9/portal/pws?t=li",
		createFormReader(formDate),
		nil,
		header,
	)
	return successFlag
}

func isConnect() bool {
	c := colly.NewCollector()
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
	for {
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

func decodeRespInfo(msg string) string {
	temp,_ := base64.StdEncoding.DecodeString(msg)
	if string(temp) != "" {
		respJson, _ := url.QueryUnescape(string(temp))
		if respJson != "" {
			 return respJson
		}
	}
	temp,_ = base64.StdEncoding.DecodeString(msg+"=")
	if string(temp) != "" {
		respJson, _ := url.QueryUnescape(string(temp))
		if respJson != "" {
			return respJson
		}
	}
	temp,_ = base64.StdEncoding.DecodeString(msg+"==")
	if string(temp) != "" {
		respJson, _ := url.QueryUnescape(string(temp))
		if respJson != "" {
			return respJson
		}
	}
	return "null"
}