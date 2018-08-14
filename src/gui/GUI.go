package gui

import (
	"github.com/mattn/go-gtk/gtk"
	"os"
	"fmt"
)

func Run() {
	gtk.Init(&os.Args) //环境初始化

	// 主窗口
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL) 		//创建窗口
	window.SetPosition(gtk.WIN_POS_CENTER)       		//设置窗口居中显示
	window.SetTitle("iMC-Login By Ericwyn") 		//设置标题
	window.SetSizeRequest(500, 400)  	//设置窗口的宽度和高度

	// 新建 layout GtkFixed
	layout := gtk.NewFixed() //创建固定布局

	//创建按钮
	b1 := newBtn("上线")
	addBtnPressHandle(b1, func() {
		fmt.Println("上线了")
	})
	b2 := newBtn("下线")
	addBtnPressHandle(b2, func() {
		fmt.Println("下线了")
	})
	b3 := newBtn("强制下线")
	addBtnPressHandle(b3, func() {
		fmt.Println("强制下线了")
	})

	// 添加 layout
	window.Add(layout) //把布局添加到主窗口中

	layout.Put(b1, 60, 320)    //设置按钮在容器的位置
	layout.Put(b2, 190, 320)
	layout.Put(b3, 320,320)

	window.ShowAll() //显示所有的控件

	gtk.Main() //主事件循环，等待用户操作
}

func newBtn(label string) *gtk.Button {
	btn := gtk.NewButton()
	btn.SetLabel(label)
	btn.SetSizeRequest(DefaultBtnWidth, DefalutBtnHeight)
	return btn
}

func addBtnPressHandle(btn *gtk.Button,f interface{}, datas ...interface{}) {
	btn.Connect("pressed", f, datas)
}