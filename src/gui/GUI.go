package gui

import (
	"github.com/mattn/go-gtk/gtk"
	"github.com/mattn/go-gtk/glib"
	"fmt"
	"utils"
)

func Run(account string, password string) {
	var menuitem *gtk.MenuItem

	gtk.Init(nil)

	// 窗口设置
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	// 标题设置
	window.SetTitle("iMCLogin")
	// 图标设置
	window.SetIconName("gtk-dialog-info")
	// 设置关闭时间
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		println("软件退出", ctx.Data().(string))
		gtk.MainQuit()
	}, "datas -> 正常退出")

	//--------------------------------------------------------
	// GtkVBox 新建布局
	//--------------------------------------------------------
	vbox := gtk.NewVBox(false, 1)

	//--------------------------------------------------------
	// GtkMenuBar 菜单
	//--------------------------------------------------------
	menubar := gtk.NewMenuBar()
	vbox.PackStart(menubar, false, false, 0)

	//--------------------------------------------------------
	// GtkVPaned
	//--------------------------------------------------------
	vpaned := gtk.NewVPaned()
	vbox.Add(vpaned)

	//--------------------------------------------------------
	// GtkFrame 页面可以分开成多个 Frame，中间有分割线来调整 Frame 的大小
	//--------------------------------------------------------
	RootFrame := gtk.NewFrame("")
	framebox := gtk.NewVBox(false, 1)
	RootFrame.Add(framebox)

	vpaned.Pack1(RootFrame, false, false)

	// 底部栏状态显示
	var statusBar *gtk.Statusbar
	var contextId uint

	// Label
	label := gtk.NewLabel("iMC Login Tool")
	label.ModifyFontEasy("DejaVu Serif 15")
	framebox.PackStart(label, false, true, 40)

	// 帐号输入框
	labelText := "请输入帐号"
	for i := 0; i < 1000; i++ {
		labelText = labelText+" "
	}
	accountLabel := gtk.NewLabel(labelText)
	accountLabel.ModifyFontEasy("12")
	accountLabel.SetPadding(0,0)
	framebox.PackStart(accountLabel, false, true, 0)
	accountEntry := gtk.NewEntry()
	accountEntry.SetText(account)
	framebox.Add(accountEntry)

	// 密码输入框
	labelText = "请输入密码"
	for i := 0; i < 1000; i++ {
		labelText = labelText+" "
	}
	pwLabel := gtk.NewLabel(labelText)
	pwLabel.ModifyFontEasy("12")

	framebox.PackStart(pwLabel, false, true, 0)
	passwordEntry := gtk.NewEntry()
	passwordEntry.SetText(password)
	framebox.Add(passwordEntry)

	buttons := gtk.NewHBox(false, 20)

	//--------------------------------------------------------
	// GtkButton 横向按钮
	//--------------------------------------------------------
	loginBtn := gtk.NewButtonWithLabel("上线")
	loginBtn.Clicked(func() {
		// TODO 异步回调
		println("按钮被点击:", loginBtn.GetLabel())
		statusBar.Push(contextId, "状态 : 正在登录")
		utils.GetCookieAndPLWithCB(func(cookie string, pl string) {
			if utils.LoginWithCB(account, password, cookie, pl, func(info string) {
				statusBar.Push(contextId, "状态 : "+info)
			}) {
				statusBar.Push(contextId, "状态 : 登录成功")
				if utils.IsConnect() {
					statusBar.Push(contextId, "状态 : 联网成功")
					go utils.StayConnectWithCB(
						account,
						password,
						cookie,
						pl,
						func(info string) {
							statusBar.Push(contextId, "状态 : "+info)
						},
					)
				}else {
					utils.Log("联网失败, 请检查网络连接")
				}
			}else {
				statusBar.Push(contextId, "状态 : 登录失败")
			}
		})
	})
	buttons.Add(loginBtn)

	logoutBtn := gtk.NewButtonWithLabel("下线")
	logoutBtn.Clicked(func() {
		statusBar.Push(contextId, "状态 : 下线按钮被点击")

		// TODO 异步回调完成下线
	})
	buttons.Add(logoutBtn)

	ensureLogoutBtn := gtk.NewButtonWithLabel("强制下线")
	ensureLogoutBtn.Clicked(func() {
		statusBar.Push(contextId, "状态 : 强制下线按钮被点击")

		// TODO 异步回调完成强制下线
	})
	buttons.Add(ensureLogoutBtn)



	framebox.PackStart(buttons, false, false, 40)

	// 菜单
	//--------------------------------------------------------
	// GtkMenuItem
	//--------------------------------------------------------
	cascademenu := gtk.NewMenuItemWithMnemonic("文件")
	menubar.Append(cascademenu)
	submenu := gtk.NewMenu()
	cascademenu.SetSubmenu(submenu)

	menuitem = gtk.NewMenuItemWithMnemonic("退出")
	menuitem.Connect("activate", func() {
		fmt.Println("通过菜单退出")
		gtk.MainQuit()
	})
	submenu.Append(menuitem)

	cascademenu = gtk.NewMenuItemWithMnemonic("关于")
	menubar.Append(cascademenu)
	submenu = gtk.NewMenu()
	cascademenu.SetSubmenu(submenu)

	menuitem = gtk.NewMenuItemWithMnemonic("作者")
	menuitem.Connect("activate", func() {
		about :=
			"开发团队 : Besfim ( https://github.com/Besfim ) \n" +
			"开发者   : Ericwyn ( https://github.com/Ericwyn ) \n\n\n"+
			"项目地址 : https://github.com/Besfim/iMC-Portal-Login"
		messagedialog := gtk.NewMessageDialog(
			menuitem.GetTopLevelAsWindow(),
			gtk.DIALOG_MODAL,
			gtk.MESSAGE_INFO,
			gtk.BUTTONS_OK,
			about)
		messagedialog.Response(func() {
			messagedialog.Destroy()
		})
		messagedialog.Run()
		//gtk.MainQuit()
	})
	submenu.Append(menuitem)

	statusBar = gtk.NewStatusbar()
	contextId = statusBar.GetContextId("go-gtk")
	statusBar.Push(contextId, "状态 : 未启动")
	framebox.PackStart(statusBar, false, false, 0)

	window.Add(vbox)
	window.SetSizeRequest(500, 400)
	window.ShowAll()
	gtk.Main()
}