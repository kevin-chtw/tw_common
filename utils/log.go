package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
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
	if writer, err := getWriter(); err != nil {
		logrus.Fatalf("Failed to create log writer: %v", err)
	} else {
		l.SetOutput(writer)
	}

	l.Formatter = &Formatter{}
	l.SetLevel(level)
	return logruswrapper.NewWithFieldLogger(l)
}

func getWriter() (*SafeRotateLogs, error) {
	// 获取程序名
	programName := filepath.Base(os.Args[0])

	// 设置日志文件路径
	logPath := "./logs"
	logFile := filepath.Join(logPath, fmt.Sprintf("%s-%%Y%%m%%d.log", programName))
	// 确保日志目录存在
	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		logrus.Fatalf("Failed to create log directory: %v", err)
	}

	// 创建日志轮转写入器
	writer, err := rotatelogs.New(
		logFile,
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		return nil, err
	}
	return &SafeRotateLogs{
		RotateLogs: writer,
		logPattern: logFile,
		maxAge:     7 * 24 * time.Hour,
		rotation:   24 * time.Hour,
	}, nil
}

// SafeRotateLogs 是一个包装器，确保文件存在
type SafeRotateLogs struct {
	*rotatelogs.RotateLogs
	logPattern string
	maxAge     time.Duration
	rotation   time.Duration
}

// Write 检查文件是否存在，如果不存在则重新创建
func (s *SafeRotateLogs) Write(p []byte) (n int, err error) {
	// 获取当前日志文件名
	currentLogFile := s.RotateLogs.CurrentFileName()

	// 检查文件是否存在
	if _, err := os.Stat(currentLogFile); os.IsNotExist(err) {
		// 如果文件不存在，重新创建日志轮转写入器
		writer, err := rotatelogs.New(
			s.logPattern,
			rotatelogs.WithMaxAge(s.maxAge),
			rotatelogs.WithRotationTime(s.rotation),
		)
		if err != nil {
			return 0, fmt.Errorf("failed to recreate log writer: %v", err)
		}
		s.RotateLogs = writer
	}

	// 写入日志
	return s.RotateLogs.Write(p)
}
