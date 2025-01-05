package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"runtime"
	"strings"
	"sync"

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
	driver  *logrus.Logger
	message interface{}
	mu      sync.Mutex // Mutex for thread safety
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
	logger.mu.Lock() // Lock for thread safety
	defer logger.mu.Unlock()

	logger.message = map[string]interface{}{
		"caller": fmt.Sprintf("%s:%d", filename, line),
		"log":    logs,
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
		l.driver.Error("Failed to marshal log message: " + err.Error())
		return // Return without panicking
	}

	logger.message = string(message)
}

func (l *logDriver) Debug() {
	l.mu.Lock() // Lock for thread safety
	defer l.mu.Unlock()
	if logger.message != nil {
		l.driver.Debug(logger.message)
		logger.message = nil
	}
}

func (l *logDriver) Info() {
	l.mu.Lock() // Lock for thread safety
	defer l.mu.Unlock()
	if logger.message != nil {
		l.driver.Info(logger.message)
		logger.message = nil
	}
}

func (l *logDriver) Warn() {
	l.mu.Lock() // Lock for thread safety
	defer l.mu.Unlock()
	if logger.message != nil {
		l.driver.Warn(logger.message)
		logger.message = nil
	}
}

func (l *logDriver) Error() {
	l.mu.Lock() // Lock for thread safety
	defer l.mu.Unlock()
	if logger.message != nil {
		l.driver.Error(logger.message)
		logger.message = nil
	}
}

func (l *logDriver) Fatal() {
	l.mu.Lock() // Lock for thread safety
	defer l.mu.Unlock()
	if logger.message != nil {
		l.driver.Fatal(logger.message)
		logger.message = nil
	}
}
