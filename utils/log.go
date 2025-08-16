package utils

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/v3/pkg/logger/interfaces"
	logruswrapper "github.com/topfreegames/pitaya/v3/pkg/logger/logrus"
)

type Formatter struct{}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(7) // 调整层级以获取正确的调用者信息
	if !ok {
		file = "???"
		line = 0
	}

	// 提取文件名
	fileName := strings.Split(file, "/")[len(strings.Split(file, "/"))-1]

	// 获取函数名
	pc, _, _, ok := runtime.Caller(7) // 调整层级以获取正确的调用者信息
	if !ok {
		pc = 0
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = strings.Split(funcName, ".")[len(strings.Split(funcName, "."))-1]

	timestamp := entry.Time.Format(time.DateTime)
	level := strings.ToLower(entry.Level.String())

	// 格式化日志
	logMessage := fmt.Sprintf("%s [%s] %s:%d %s %s\n", timestamp, level, fileName, line, funcName, entry.Message)

	return []byte(logMessage), nil
}

func Logger(level logrus.Level) interfaces.Logger {
	l := logrus.New()
	l.Formatter = &Formatter{}
	l.SetLevel(level)
	return logruswrapper.NewWithFieldLogger(l)
}
