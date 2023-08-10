package main

import (
	"dirHelper/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// 提示需要跳转的程序exit码
const jumpExitCode = 10
const dataDirName = ".dirHelper"
const rcName = ".dirHelperrc"
const dataFilename = "dirHelper.json"
const importRcContent = "source ~/" + dataDirName + "/" + rcName

var userRc = []string{".bashrc", ".zshrc"}

// 从用户目录中读取dirHelper的数据
func readData() *Data {
	var data Data
	data.DirMap = make(map[string]string)

	jsonData, err := ioutil.ReadFile(path.Join(getDataDir(), dataFilename))
	if err != nil {
		return &data
	}
	_ = json.Unmarshal(jsonData, &data)
	return &data
}

// 保存数据
func writeData(data *Data) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return utils.Logger.Errorf("json marshal err[%s]", err)
	}

	err = ioutil.WriteFile(path.Join(getDataDir(), dataFilename), jsonData, 0644)
	if err != nil {
		return utils.Logger.Errorf("write data file err[%s]", err)
	}

	return nil
}

// 获取数据的存放目录
func getDataDir() string {
	currentUser, _ := user.Current()
	return currentUser.HomeDir + "/" + dataDirName
}

// 初始化shell相关环境
func initShell(name string) error {

	// 获取当前程序的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		return utils.Logger.Errorf("get program path err[%s]", err)
	}
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return utils.Logger.Errorf("get program abs path err[%s]", err)
	}

	// 创建rc文件
	content := `
# 设置tab键补全的时候 不区分大小写
bind "set completion-ignore-case on"

# 执行dirHelper 并根据返回结果
dirHelperFun() {
    # 执行命令并获取结果
    res="` + "`" + `` + absPath + ` $@` + "`" + `"
    # 如果命令返回10 说明是要跳转目录 且返回结果为待跳转的目录
    if [ $? -eq ` + strconv.Itoa(jumpExitCode) + ` ]; then
        cd $res
    else
        echo "$res"
    fi
}

# 添加命令定义
alias ` + name + `="dirHelperFun"

`

	err = ioutil.WriteFile(path.Join(getDataDir(), rcName), []byte(content), 0644)
	if err != nil {
		return utils.Logger.Errorf("write rc file err[%s]", err)
	}

	// 在用户的rc文件中包含dirHelper的rc文件
	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir
	for i := range userRc {
		rcFile := path.Join(homeDir, userRc[i])
		fileData, err := ioutil.ReadFile(rcFile)
		if err != nil {
			continue
		}
		fileContent := string(fileData)
		if !strings.Contains(fileContent, importRcContent) {
			fileContent += "\n" + importRcContent + "\n"
			err = ioutil.WriteFile(rcFile, []byte(fileContent), 0644)
			if err != nil {
				return utils.Logger.Errorf("write rc file err[%s]", err)
			}
		}
	}

	return nil
}

// 添加目录
func addDir(dir string) error {
	data := readData()

	// 以数字作为key
	num := 0
	numStr := strconv.Itoa(num)
	exist := true

	// key不能重复
	for exist {
		num++
		numStr = strconv.Itoa(num)
		_, exist = data.DirMap[numStr]
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return utils.Logger.Errorf("get dir[%s] abs path err[%s]", dir, err)
	}
	data.DirMap[numStr] = absDir

	// 写入数据
	err = writeData(data)
	if err != nil {
		return err
	}

	return nil
}

func showDir() {
	data := readData()

	num := len(data.DirMap)
	colorPrint(fmt.Sprintf("total num[%d], below is dir list:\n", num))

	for k, v := range data.DirMap {
		colorPrint(fmt.Sprintf("%-10s %s\n", k, v))
	}
}

func rmDir(key string) error {
	data := readData()

	_, ok := data.DirMap[key]
	if !ok {
		return utils.Logger.Errorf("this key[%s] not in dir list", key)
	}
	delete(data.DirMap, key)

	// 写入数据
	err := writeData(data)
	if err != nil {
		return err
	}
	return nil
}

func chgDir(param string) error {
	data := readData()

	res := strings.Split(param, ",")
	if len(res) != 2 {
		return utils.Logger.Errorf("param[%s] err", param)
	}
	oldKey := res[0]
	newKey := res[1]

	v, ok := data.DirMap[oldKey]
	if !ok {
		return utils.Logger.Errorf("this key[%s] not in dir list", oldKey)
	}
	delete(data.DirMap, oldKey)
	data.DirMap[newKey] = v

	// 写入数据
	err := writeData(data)
	if err != nil {
		return err
	}
	return nil
}

func getDirByKey(key string) (string, error) {
	data := readData()

	v, ok := data.DirMap[key]
	if !ok {
		return "", utils.Logger.Errorf("this key[%s] not in dir list", key)
	}

	return v, nil
}

func colorPrint(str string) {
	fmt.Printf("\033[1;32m%s\033[0m", str)
}