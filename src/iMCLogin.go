package main

import (
	"fmt"
		"strings"
		"flag"
	"os"
	"bufio"
	"utils"
	"gui"
)

var Debug = flag.Bool("d",false,"开启 DEBUG 日志打印")
// 强制退出
var EnsureLogoutFlag = flag.Bool("o",false,"强制退出特定帐号, 也要求正确的用户名和密码")
// 学号
var UserNum = flag.String("u","","学生学号")
// 密码
var UserPw = flag.String("p","","校园网密码")

var UserCommonLine = flag.Bool("c",false,"使用命令行模式运行")



func main() {
	initImcGDPU()
	if *UserCommonLine {
		CommonLineRun()
	}else {
		GuiRun(*UserNum, *UserPw)
	}
}

func CommonLineRun() {
	utils.DebugLog("userName:" + *UserNum + ", password: " + *UserPw)

	if *UserNum == "" || *UserPw == "" {
		fmt.Println("用户名或密码为空, 或参数错误, 请使用 -h 查看使用帮助")
		os.Exit(0)
	}else {
		// colly.Async(false)
		// TODO 作者标记
		// TODO 获取新版本
	}

	cookie,pl := utils.GetCookieAndPL()
	if cookie == "null" && pl == "null" {
		fmt.Println("程序退出")
		os.Exit(0)
	}
	if utils.Login(*UserNum, *UserPw, cookie, pl) {
		if utils.IsConnect() {
			go utils.StayConnect(*UserNum, *UserPw, cookie, pl)
			utils.Log("联网成功, 输入 exit 可下线并退出程序")
			go inputExit(cookie, pl)
			select {}
		}else {
			utils.Log("联网失败, 请检查网络连接")
		}
	}
}

func GuiRun(account string, password string) {
	gui.Run(account,password)
}

func inputExit(cookie string, pl string ){
	for {
		inputReader := bufio.NewReader(os.Stdin)
		input, err := inputReader.ReadString('\n')
		if err != nil {
			utils.DebugLog("获取用户输入失败, : " + err.Error())
			utils.DebugLog("请停止程序之后使用 -o 强制退出, 确保帐号登出")
			break
		}
		if strings.Contains(input,"exit") { // Windows, on Linux it is "S\n"
			if utils.Logout(*UserNum, cookie,pl) {
				utils.Log("下线成功")
			}else {
				utils.DebugLog("下线失败, 使用强制下线")
				if utils.EnsureLogout(*UserNum, *UserPw) {
					utils.Log("下线成功")
				}else {
					utils.Log("下线失败")
				}
			}
			utils.Log("程序已退出")
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
		utils.DebugMode = true
	}
	if *EnsureLogoutFlag {
		if utils.EnsureLogout(*UserNum, *UserPw) {
			utils.Log("退出成功")
		}else {
			utils.Log("退出失败, 请使用 -d 开启 Debug 模式查看日志")
		}
		os.Exit(0)
	}
}