package delayQ

import (
	"fmt"
	"time"
)

// todo: 添加日志级别的选项，由使用者自己配置
// 如果没有设置Logger，则选用默认Logger
var defaultLogger = new(printfLogger)

type Logger interface {
	InfoF(format string, args ...interface{})
	ErrorF(format string, args ...interface{})
}

type printfLogger struct{}

func (printfLogger) InfoF(format string, args ...interface{}) {
	fmt.Printf("[INFO] - %s - %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
}

func (printfLogger) ErrorF(format string, args ...interface{}) {
	fmt.Printf("[\033[31mERROR\033[0m] - %s - [\033[31m%s\033[0m]\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
}
