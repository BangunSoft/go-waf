package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
	PANIC = "panic"
	FATAL = "fatal"
)

type logDriver struct {
	driver *logrus.Logger

	message interface{}
}

var logger = &logDriver{
	driver: logrus.New(),
}

func SetLevel(level string) {
	level = strings.ToLower(level)

	switch level {
	case DEBUG:
		logger.driver.Level = logrus.DebugLevel
	case INFO:
		logger.driver.Level = logrus.InfoLevel
	case WARN:
		logger.driver.Level = logrus.WarnLevel
	case ERROR:
		logger.driver.Level = logrus.ErrorLevel
	case PANIC:
		logger.driver.Level = logrus.PanicLevel
	case FATAL:
		logger.driver.Level = logrus.FatalLevel
	default:
		logger.driver.Level = logrus.GetLevel()
	}
}

func SetOutput(output io.Writer) {
	log.Println("[info] logger output is changed")
	logger.driver.SetOutput(output)
}

func Logger(logs ...interface{}) *logDriver {
	if len(logs) == 0 || logs[0] == nil {
		return logger
	}

	_, filename, line, _ := runtime.Caller(1)
	logger.message = map[string]interface{}{
		"caller": fmt.Sprintf("%s:%d", filename, line),
		"log":    fmt.Sprintf("%v", logs),
	}
	logger.jsonize()

	return logger
}

func (l *logDriver) jsonize() {
	if logger.message == nil {
		return
	}

	message, err := json.Marshal(logger.message)
	if err != nil {
		log.Panicln("[panic] logger error. Causer: ", err)
	}

	logger.message = string(message)
}

func (l *logDriver) Debug() {
	if logger.message != nil {
		l.driver.Debug(l.message)
		l.message = nil
	}
}

func (l *logDriver) Info() {
	if logger.message != nil {
		l.driver.Info(l.message)
		l.message = nil
	}
}

func (l *logDriver) Warn() {
	if logger.message != nil {
		l.driver.Warn(l.message)
		l.message = nil
	}
}

func (l *logDriver) Error() {
	if logger.message != nil {
		l.driver.Error(l.message)
		l.message = nil
	}
}

func (l *logDriver) Fatal() {
	if logger.message != nil {
		l.driver.Fatal(l.message)
		l.message = nil
	}
}
