# iMC-Portal-Login
适用于 GDPU 的 iMC Portal 无线网络登录工具, 

## 前言说明
学校之前开通了 iMC Portal 无线网络, 只是每次都要开启浏览器才可以登录, 并且需要一直开着一个标签, 这是件很麻烦的事情, 而且最近 (2018-07-11) 发现会出现每隔 15 分钟就断线一次的问题, 所以抓包来模拟登录, 并且在断线之后自动重连
 
 程序用 Golang 来编码所以可以方便的打包成二进制文件给各个平台使用

## 使用教程
## 编译 

    go build iMCLogin.go -o iMCLogin

## 启动

    ./iMCLogin -u 学号 -p 用户密码

## 参数列表

    -d
        开启 DEBUG 日志打印
    -p string
        校园网密码 (default "null")
    -u string
        学生学号 (default "null")


