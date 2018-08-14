package web

import (
	"fmt"
	"time"
)

var DebugMode bool = false

// 打印 debug 日志
func DebugLog(msg string) {
	if DebugMode {
		fmt.Println("[LOG] [DEBUG]   ["+getTime(time.Now())+"]:", msg)
	}
}

func Log(msg string) {
	fmt.Println("[LOG] [RUNNING] ["+getTime(time.Now())+"]:",msg)
}

func getTime(time time.Time) string {
	return time.String()[5:19]
}