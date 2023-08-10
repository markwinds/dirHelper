package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

const (
	DebugLevel = 0
	InfoLevel  = 1
	WarnLevel  = 2
	ErrorLevel = 3
)

const baseDeep = 2

var levelStrList = []string{"D", "I", "W", "E"}

type logPack struct {
	content string // 日志内容
	level   int    // 日志等级
}

type SimpleLogger struct {
	Level    int           // 日志记录等级
	Color    bool          // 打印是否有颜色
	Stack    bool          // 是否打印出error等级的调用栈
	Filesize int64         // 日志达到多大时进行压缩
	Dir      string        // 日志存放目录
	Filename string        // 日志文件名称
	Deep     int           // 文件名和行号调用栈的深度
	c        chan *logPack // 队列
	data     *logPack      // 当前正在写的日志数据
}

func (l *SimpleLogger) Init() error {

	l.c = make(chan *logPack, 100)
	l.data = nil

	l.Deep += baseDeep

	// 创建目录
	if l.Dir == "" {
		l.Dir = "./"
	}
	err := os.MkdirAll(l.Dir, 0700)
	if err != nil {
		fmt.Printf("mk dir err[%s]", err)
		return err
	}

	// 启动线程写
	go l.write()

	return nil
}

// Wait 等待异步日志全部写完
func (l *SimpleLogger) Wait() {
	// 已经没有数据正在写 或 管道中有数据
	for {
		if l.data == nil && len(l.c) == 0 {
			time.Sleep(time.Millisecond * time.Duration(10))
			if l.data == nil && len(l.c) == 0 { // 隔一小段时间再次判断
				break
			}
		}
		time.Sleep(time.Millisecond * time.Duration(10))
	}
}

// Debugf 格式化输出debug信息
func (l *SimpleLogger) Debugf(format string, v ...any) error {
	return l.log(DebugLevel, format, v...)
}

// Infof 格式化输出info信息
func (l *SimpleLogger) Infof(format string, v ...any) error {
	return l.log(InfoLevel, format, v...)
}

// Warnf 格式化输出warn信息
func (l *SimpleLogger) Warnf(format string, v ...any) error {
	return l.log(WarnLevel, format, v...)
}

// Errorf 格式化输出error信息
func (l *SimpleLogger) Errorf(format string, v ...any) error {
	return l.log(ErrorLevel, format, v...)
}

func (l *SimpleLogger) log(level int, format string, v ...any) (e error) {

	formatStr := fmt.Sprintf(format, v...)
	defer func() {
		e = fmt.Errorf(formatStr)
	}()

	if l.Level > level {
		return
	}

	// 组装字符串
	_, file, line, _ := runtime.Caller(l.Deep)
	content := fmt.Sprintf("[%s][%s][%s:%d]  ",
		levelStrList[level],
		time.Now().Format("2006-01-02T15:04:05.000000"),
		file,
		line)
	content += formatStr
	content += "\n"

	// 获取调用栈打印
	if l.Stack && level == ErrorLevel {
		stackStr := string(debug.Stack())
		count := 0 // 去除前面几行没用的栈打印

		//stackList := strings.Split(stackStr, "\n")[3+2*l.Deep:]
		//
		//for i := range stackList {
		//	content += stackList[i] + "\n"
		//}

		tmpStackStr := ""
		for i, _ := range stackStr {
			if stackStr[i] == '\n' {
				count++
			}
			if count == 3+2*l.Deep {
				if len(stackStr)-1 > i {
					tmpStackStr += stackStr[i+1:]
					break
				}
			}
		}
		tmpStrList := strings.Split(tmpStackStr, "\n")
		for i := range tmpStrList {
			if i%2 == 1 {
				content += tmpStrList[i] + "\n"
			}
		}

		//pc, file, line, _ := runtime.Caller(l.Deep)
		//pcName := runtime.FuncForPC(pc).Name()
		//content += file + ":" + strconv.Itoa(line) + " " + pcName+"\n"
	}

	// 发送
	l.c <- &logPack{
		content: content + "\n",
		level:   level,
	}

	return
}

func (l *SimpleLogger) write() {

	logPath := path.Join(l.Dir, l.Filename)

	for {
		// 更新数据为nil 作为判断是否有日志正在写的依据
		l.data = nil

		l.data = <-l.c

		// 打印
		if l.Color {
			switch l.data.level {
			case DebugLevel:
				fmt.Printf("\033[1;34m%s\033[0m", l.data.content)
			case InfoLevel:
				fmt.Printf("\033[1;32m%s\033[0m", l.data.content)
			case WarnLevel:
				fmt.Printf("\033[1;33m%s\033[0m", l.data.content)
			case ErrorLevel:
				fmt.Printf("\033[1;31m%s\033[0m", l.data.content)
			}
		} else {
			fmt.Printf("%s", l.data.content)
		}

		// 保存到文档
		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("open file[%s] to log err[%s]", logPath, err)
			continue
		}
		_, err = f.Write([]byte(l.data.content))
		if err != nil {
			fmt.Printf("write file err[%s]", err)
			_ = f.Close()
			continue
		}
		stat, err := f.Stat()
		_ = f.Close()
		if err != nil {
			fmt.Printf("get file[%s] stat err[%s]", logPath, err)
			continue
		}

		// 压缩
		if stat.Size() < l.Filesize {
			continue
		}
		compressFilename := time.Now().Format("20060102150405") + ".tar.gz"

		// 压缩日志不能调使用了log的函数，因为日志队列如果满了就会卡死在压缩
		d, _ := os.Create(path.Join(l.Dir, compressFilename))
		gw := gzip.NewWriter(d)
		tw := tar.NewWriter(gw)
		f, _ = os.Open(logPath)
		info, _ := f.Stat()
		header, _ := tar.FileInfoHeader(info, "")
		header.Name = logPath
		_ = tw.WriteHeader(header)
		_, _ = io.Copy(tw, f)
		f.Close()
		tw.Close()
		gw.Close()
		d.Close()

		_ = os.Remove(logPath)
	}
}
