package web


import (
	"github.com/gocolly/colly"
	"net/http"
	"os"
	"strings"
	"encoding/base64"
	"net/url"
	"strconv"
	"time"
	"io"
)

// 起始联网时间戳
var startTime string

//获取 cookie
func GetCookieAndPL() (string, string) {
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
		DebugLog("登录页面返回 cookie : "+cookie)
		Log("登录页面访问失败, 程序退出")
		os.Exit(0)
	}else {
		pl = strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	}
	return cookie, pl
}

//登录
func Login(userNum string, userPw string, cookie string, pl string) bool {
	c := colly.NewCollector()
	successFlag := false

	header := http.Header{
		"Cookie":          []string{cookie},
		"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7"},
	}
	formDate := map[string]string{
		"userName":            userNum,
		"userPwd":             base64.StdEncoding.EncodeToString([]byte(userPw)),
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
		DebugLog("登录返回info : " + info)
		decodeTemp := decodeRespInfo(info)
		respJson, _ := url.QueryUnescape(decodeTemp)
		DebugLog("登录返回 : " + respJson)
		startTime = strconv.FormatInt(time.Now().UnixNano(),10)[0:13]
		// 登录成功 {"errorNumber":"1","heartBeatCyc":900000,"heartBeatTimeoutMaxTime":2,"userDevPort":"Servers-Aggregation-SW-S5560X-54C-EI-vlan-01-4055@vlan","userStatus":99,"serialNo":4686,"ifNeedModifyPwd":false,"browserUrl":"","clientPrivateIp":"","userurl":"","usermac":null,"nasIp":"","clientLanguage":"Chinese","ifTryUsePopupWindow":true,"triggerRedirectUrl":"","portalLink":"JTdCJTIyZXJyb3JOdW1iZXIlMjIlM0ElMjIxJTIyJTJDJTIyaGVhcnRCZWF0Q3ljJTIyJTNBOTAwMDAwJTJDJTIyaGVhcnRCZWF0VGltZW91dE1heFRpbWUlMjIlM0EyJTJDJTIydXNlckRldlBvcnQlMjIlM0ElMjJTZXJ2ZXJzLUFnZ3JlZ2F0aW9uLVNXLVM1NTYwWC01NEMtRUktdmxhbi0wMS00MDU1JTQwdmxhbiUyMiUyQyUyMnVzZXJTdGF0dXMlMjIlM0E5OSUyQyUyMnNlcmlhbE5vJTIyJTNBNDY4NiUyQyUyMmlmTmVlZE1vZGlmeVB3ZCUyMiUzQWZhbHNlJTJDJTIyYnJvd3NlclVybCUyMiUzQSUyMiUyMiUyQyUyMmNsaWVudFByaXZhdGVJcCUyMiUzQSUyMiUyMiUyQyUyMnVzZXJ1cmwlMjIlM0ElMjIlMjIlMkMlMjJ1c2VybWFjJTIyJTNBbnVsbCUyQyUyMm5hc0lwJTIyJTNBJTIyJTIyJTJDJTIyY2xpZW50TGFuZ3VhZ2UlMjIlM0ElMjJDaGluZXNlJTIyJTJDJTIyaWZUcnlVc2VQb3B1cFdpbmRvdyUyMiUzQXRydWUlMkMlMjJ0cmlnZ2VyUmVkaXJlY3RVcmwlMjIlM0ElMjIlMjIlN0Q"}
		if strings.Contains(respJson, "errorNumber") && strings.Contains(respJson, "heartBeatTimeoutMaxTime") {
			successFlag = true
		}else if strings.Contains(respJson,"E63032:密码错误") {
			// 密码错误 	{"portServIncludeFailedCode":"63032","portServIncludeFailedReason":"E63032:密码错误，您还可以重试8次。","e_c":"portServIncludeFailedCode","e_d":"portServIncludeFailedReason","errorNumber":"7"}
			Log("用户密码错误")
			successFlag = false
		}else if strings.Contains(respJson,"E63018:用户不存在或者用户没有申请该服务") {
			// 用户名不存在 {"portServIncludeFailedCode":"63018","portServIncludeFailedReason":"E63018:用户不存在或者用户没有申请该服务。","e_c":"portServIncludeFailedCode","e_d":"portServIncludeFailedReason","errorNumber":"7"}
			Log("用户已经登录, 尝试强制下线后重新登录")
			successFlag = false
		}else if strings.Contains(respJson,"设备拒绝请求") {
			// 已经登录		{"portServErrorCode":"1","portServErrorCodeDesc":"设备拒绝请求","e_c":"portServErrorCode","e_d":"portServErrorCodeDesc","errorNumber":"7"}
			Log("用户已经登录, 尝试强制下线后重新登录")
			successFlag = Login(userNum, userPw, cookie, pl)
		}else {
			//
			Log("登录失败, 错误信息如下: "+respJson)
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

func HeartBeat(cookie string, pl string) {
	c := colly.NewCollector()
	c.OnResponse(func(response *colly.Response) {
		DebugLog("成功发送了一个心跳包")
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

func Logout(userNum string, cookie string, pl string) bool {
	c := colly.NewCollector()
	resFlag := false
	c.OnResponse(func(response *colly.Response) {
		DebugLog("成功发送了退出请求")
		info := string(response.Body)
		decodeTemp := decodeRespInfo(info)
		respJson, _ := url.QueryUnescape(decodeTemp)
		// 如果可以可以 GET 这个地址并且得到返回的话, 就证明之前是在线的, 并且现在退出了
		// 返回的 json 是 {"errorNumber":"1"}
		DebugLog(respJson)
		if strings.Contains(respJson,"errorNumber") {
			DebugLog("退出成功")
			resFlag = true
		}
	})
	header := http.Header{
		"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7"},
		"Cookie":		   []string{"hello1="+userNum+"; hello2=false;"+cookie},
		"Referer":         []string{"http://10.50.15.9/portal/page/online.jsp?st=2&pl=" + pl + "&custompath=templatePage/20170110154814101/&uamInitCustom=0&customCfg=MTA1&uamInitLogo=H3C&userName=null&userPwd=null&loginType=3&innerStr=null&outerStr=null&v_is_selfLogin=0"},
	}
	c.Request("GET", "http://10.50.15.9/portal/pws?t=lo&language=Chinese&userip=&basip=&_="+startTime, nil, nil, header)
	return resFlag
}

//这个方法是用来确保程序已经退出的
func EnsureLogout(userNum string, userPw string) bool {
	cookie,pl := GetCookieAndPL()
	c := colly.NewCollector()
	successFlag := false

	header := http.Header{
		"Cookie":          []string{cookie},
		"User-Agent":      []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36"},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9,en-GB;q=0.8,en;q=0.7"},
	}
	formDate := map[string]string{
		"userName":            userNum,
		"userPwd":             base64.StdEncoding.EncodeToString([]byte(userPw)),
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
		DebugLog("登录返回info : " + info)
		decodeTemp := decodeRespInfo(info)
		respJson, _ := url.QueryUnescape(decodeTemp)
		DebugLog("登录返回 : " + respJson)
		startTime = strconv.FormatInt(time.Now().UnixNano(),10)[0:13]
		// 登录成功 {"errorNumber":"1","heartBeatCyc":900000,"heartBeatTimeoutMaxTime":2,"userDevPort":"Servers-Aggregation-SW-S5560X-54C-EI-vlan-01-4055@vlan","userStatus":99,"serialNo":4686,"ifNeedModifyPwd":false,"browserUrl":"","clientPrivateIp":"","userurl":"","usermac":null,"nasIp":"","clientLanguage":"Chinese","ifTryUsePopupWindow":true,"triggerRedirectUrl":"","portalLink":"JTdCJTIyZXJyb3JOdW1iZXIlMjIlM0ElMjIxJTIyJTJDJTIyaGVhcnRCZWF0Q3ljJTIyJTNBOTAwMDAwJTJDJTIyaGVhcnRCZWF0VGltZW91dE1heFRpbWUlMjIlM0EyJTJDJTIydXNlckRldlBvcnQlMjIlM0ElMjJTZXJ2ZXJzLUFnZ3JlZ2F0aW9uLVNXLVM1NTYwWC01NEMtRUktdmxhbi0wMS00MDU1JTQwdmxhbiUyMiUyQyUyMnVzZXJTdGF0dXMlMjIlM0E5OSUyQyUyMnNlcmlhbE5vJTIyJTNBNDY4NiUyQyUyMmlmTmVlZE1vZGlmeVB3ZCUyMiUzQWZhbHNlJTJDJTIyYnJvd3NlclVybCUyMiUzQSUyMiUyMiUyQyUyMmNsaWVudFByaXZhdGVJcCUyMiUzQSUyMiUyMiUyQyUyMnVzZXJ1cmwlMjIlM0ElMjIlMjIlMkMlMjJ1c2VybWFjJTIyJTNBbnVsbCUyQyUyMm5hc0lwJTIyJTNBJTIyJTIyJTJDJTIyY2xpZW50TGFuZ3VhZ2UlMjIlM0ElMjJDaGluZXNlJTIyJTJDJTIyaWZUcnlVc2VQb3B1cFdpbmRvdyUyMiUzQXRydWUlMkMlMjJ0cmlnZ2VyUmVkaXJlY3RVcmwlMjIlM0ElMjIlMjIlN0Q"}
		if strings.Contains(respJson,"errorNumber") && strings.Contains(respJson, "heartBeatTimeoutMaxTime") {
			Log("尝试登录成功")
			// 似乎这里的 logout 一直是没法使用的
			tempFlag := Logout(userNum, cookie, pl)
			if tempFlag {
				DebugLog("下线成功")
				successFlag = true
			}else {
				DebugLog("正常下线失败, 尝试重新强制下线")
				//如果这个时候下线不了的话就来多一次登录
				successFlag = EnsureLogout(userNum, userPw)
			}
		}else if strings.Contains(respJson,"设备拒绝请求") {
			DebugLog("强制下线成功")
			successFlag = true
		}else {
			//
			Log("强制下线失败, 尝试登录后错误信息如下: "+respJson)
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

func IsConnect() bool {
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
		DebugLog("网络联通")
	} else {
		DebugLog("网络不联通")
	}
	return connectFlag
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

func createFormReader(data map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range data {
		form.Add(k, v)
	}
	return strings.NewReader(form.Encode())
}