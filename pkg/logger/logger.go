package logger

import (
	"encoding/json"
	"io"
	"log"
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

func Logger(logs ...any) *logDriver {
	if len(logs) == 0 || logs[0] == nil {
		return logger
	}

	logger.message = logs
	logger.jsonize()

	return logger
}

func (l *logDriver) jsonize() {
	if logger.message == nil {
		return
	}

	var err error
	logger.message, err = json.Marshal(logger.message)

	if err != nil {
		log.Panicln("[panic] logger error. Causer: ", err)
	}
}

func (l *logDriver) Debug() {
	if logger.message != nil {
		l.driver.Debug(l.message)
	}
}

func (l *logDriver) Info() {
	if logger.message != nil {
		l.driver.Info(l.message)
	}
}

func (l *logDriver) Warn() {
	if logger.message != nil {
		l.driver.Warn(l.message)
	}
}

func (l *logDriver) Error() {
	if logger.message != nil {
		l.driver.Error(l.message)
	}
}

func (l *logDriver) Fatal() {
	if logger.message != nil {
		l.driver.Fatal(l.message)
	}
}
