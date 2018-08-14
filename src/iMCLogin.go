package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"strings"
	"time"
	"flag"
	"os"
	"bufio"
	"web"
)

var Debug = flag.Bool("d",false,"开启 DEBUG 日志打印")
// 强制退出
var EnsureLogoutFlag = flag.Bool("o",false,"强制退出特定帐号, 也要求正确的用户名和密码")
// 学号
var UserNum = flag.String("u","null","学生学号")
// 密码
var UserPw = flag.String("p","null","校园网密码")

// 心跳包发送间隔,单位是秒
const heartBeatInterval time.Duration = 30

func main() {

	initImcGDPU()
	web.DebugLog("userName:" + *UserNum + ", password: " + *UserPw)

	cookie,pl := web.GetCookieAndPL()
	if web.Login(*UserNum, *UserPw, cookie, pl) {
		if web.IsConnect() {
			go stayConnect(cookie, pl)
			web.Log("联网成功, 输入 exit 可下线并退出程序")
			go inputExit(cookie, pl)
			select {}
		}else {
			web.Log("联网失败, 请检查网络连接")
		}
	}
}

func inputExit(cookie string, pl string ){
	for {
		inputReader := bufio.NewReader(os.Stdin)
		input, err := inputReader.ReadString('\n')
		if err != nil {
			web.DebugLog("获取用户输入失败, : " + err.Error())
			web.DebugLog("请停止程序之后使用 -o 强制退出, 确保帐号登出")
			break
		}
		if strings.Contains(input,"exit") { // Windows, on Linux it is "S\n"
			if web.Logout(*UserNum, cookie,pl) {
				web.Log("下线成功")
			}else {
				web.DebugLog("下线失败, 使用强制下线")
				if web.EnsureLogout(*UserNum, *UserPw) {
					web.Log("下线成功")
				}else {
					web.Log("下线失败")
				}
			}
			web.Log("程序已退出")
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
	if *Debug {
		web.DebugMode = true
	}
	if *EnsureLogoutFlag {
		if web.EnsureLogout(*UserNum, *UserPw) {
			web.Log("退出成功")
		}else {
			web.Log("退出失败, 请使用 -d 开启 Debug 模式查看日志")
		}
		os.Exit(0)
	}
	if *UserNum == "null" || *UserPw == "null" {
		fmt.Println("用户名或密码为空, 或参数错误, 请使用 -h 查看使用帮助")
		os.Exit(0)
	}else {
		colly.Async(false)
		// TODO 作者标记
		// TODO 获取新版本
		// TODO 强制下线的 flag
	}
}

var reConnectTime = 0
func stayConnect(cookie string, pl string)  {
	for {
		if reConnectTime > 10 {
			web.Log("连接失败次数过多, 请确认是否联网")
		}
		web.HeartBeat(cookie, pl)
		if !web.IsConnect() {
			web.Log("发送心跳包后联网失败, 尝试重新连接")
			reConnectTime++
			cookie,pl := web.GetCookieAndPL()
			web.Login(*UserNum, *UserPw, cookie,pl)
			if web.IsConnect() {
				web.Log("联网成功")
			}else {
				continue
			}
		}else {
			reConnectTime = 0
		}
		time.Sleep(time.Second * heartBeatInterval)
	}
}