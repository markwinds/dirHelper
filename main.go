package main

import (
	"dirHelper/utils"
	"flag"
	"fmt"
)

var (
	AddDir  string
	RmDir   string
	ChgDir  string
	CmdName string // 初始化dirHelper功能，创建别名
)

func init() {
	flag.StringVar(&CmdName, "i", "", "init dirHelper, appoint a command name\nexample: dirHelper -i dirHelp")
	flag.StringVar(&AddDir, "a", "", "add dir\nexample: dirHelper -a /home/mark")
	flag.StringVar(&RmDir, "r", "", "remove dir by key\nexample: dirHelper -d 1")
	flag.StringVar(&ChgDir, "c", "", "change dir key\nexample: dirHelper -c /home/mark")
}

func main() {
	flag.Parse()

	dataDir := getDataDir()

	// 初始化日志记录器
	utils.Logger = &utils.SimpleLogger{
		Level:    utils.WarnLevel,
		Color:    true,
		Stack:    false,
		Filesize: 100 << 20,
		Dir:      dataDir,
		Filename: "dirHelper.log"}
	_ = utils.Logger.Init()

	// 判断是否是初始化
	if CmdName != "" {
		err := initShell(CmdName)
		if err != nil {
			utils.Exit(1)
		}
		utils.Logger.Infof("init ok")
		return
	}

	// 添加目录信息
	if AddDir != "" {
		err := addDir(AddDir)
		if err != nil {
			utils.Exit(1)
		}
		showDir()
		return
	}

	// 删除目录
	if RmDir != "" {
		err := rmDir(RmDir)
		if err != nil {
			utils.Exit(1)
		}
		showDir()
		return
	}

	// 修改目录键值
	if ChgDir != "" {
		err := chgDir(ChgDir)
		if err != nil {
			utils.Exit(1)
		}
		showDir()
		return
	}

	// 判断是否是目录跳转逻辑
	// 如果有未被解析的参数，当作要跳转到目录
	keys := flag.Args()
	if len(keys) != 0 {
		dir, err := getDirByKey(keys[0])
		if err != nil {
			utils.Exit(1)
		}
		fmt.Println(dir)
		utils.Exit(jumpExitCode)
	}

	// 什么参数也没有 打印列表
	showDir()

}
