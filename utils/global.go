package utils

import "os"

var (
	Logger *SimpleLogger // 日志记录器
)

func Exit(code int) {
	Logger.Wait()
	os.Exit(code)
}
