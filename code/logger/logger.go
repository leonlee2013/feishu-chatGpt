package logger

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

// https://github.com/sirupsen/logrus
var logger = logrus.New()

func Log() *logrus.Logger {
	return logger
}

func WithField(key string, value interface{}) *logrus.Entry {
	return logger.WithField(key, value)
}

//logrus.WithFields(logrus.Fields{
//"animal": "walrus",
//}).Info("A walrus appears")

func init() {
	format := new(logrus.TextFormatter)
	format.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
		file := f.File
		// 将文件名从完整路径中提取
		shortFileName := file
		if idx := strings.LastIndex(file, "/"); idx != -1 {
			shortFileName = file[idx+1:]
		}
		funcVal := fmt.Sprintf("%s()", f.Function)
		fileVal := fmt.Sprintf("%s:%d", shortFileName, f.Line)
		return funcVal, fileVal

	}
	//logger.SetFormatter(&formatter{})
	logger.SetFormatter(format)

	logger.SetReportCaller(true)

	gin.DefaultWriter = logger.Out

	// 设置日志级别 支持
	//PanicLevel
	//FatalLevel
	//ErrorLevel
	//WarnLevel
	//InfoLevel
	//DebugLevel
	//logger.Level = logrus.InfoLevel
	logger.Level = logrus.DebugLevel

}

type Fields logrus.Fields

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	if logger.Level >= logrus.DebugLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Debugf(format, args...)
	}
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	if logger.Level >= logrus.WarnLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Warnf(format, args...)
	}
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	if logger.Level >= logrus.ErrorLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Errorf(format, args...)
	}
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	if logger.Level >= logrus.FatalLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Fatalf(format, args...)
	}
}

func Debug(format string, args ...interface{}) {
	if logger.Level >= logrus.DebugLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Debug(format, args)
	}
}

// Info logs a message at level Info on the standard logger.
func Info(format string, args ...interface{}) {
	if logger.Level >= logrus.InfoLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Info(format, args)
	}
}

// Warn logs a message at level Warn on the standard logger.
func Warn(format string, args ...interface{}) {
	if logger.Level >= logrus.WarnLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Warn(format, args)
	}
}

// Error logs a message at level Error on the standard logger.
func Error(format string, args ...interface{}) {
	if logger.Level >= logrus.ErrorLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Error(format, args)
	}
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(format string, args ...interface{}) {
	if logger.Level >= logrus.FatalLevel {
		entry := logger.WithFields(logrus.Fields{})
		entry.Fatal(format, args)
	}
}

// Formatter implements logrus.Formatter interface.
type formatter struct {
	prefix string
}

// Format building log message.
func (f *formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var sb bytes.Buffer

	sb.WriteString("[" + strings.ToUpper(entry.Level.String()) + "]")
	sb.WriteString(entry.Time.Format("2006-01-02 15:04:05"))
	sb.WriteString(" ")
	//sb.WriteString(" ")
	//sb.WriteString(f.prefix)
	sb.WriteString(entry.Message)
	sb.WriteString("\n")

	return sb.Bytes(), nil
}

//func WithFields(fields logrus.Fields) {
//
//}

//logrus.WithFields(logrus.Fields{
//"animal": "walrus",
//}).Info("A walrus appears")
//
//logrus.WithFields(logrus.Fields{
//"animal": "walrus",
//"size":   10,
//}).Info("A group of walrus emerges from the ocean")
//
//logrus.WithFields(logrus.Fields{
//"omg":    true,
//"number": 122,
//}).Warn("The group's number increased tremendously!")
//
//// A common pattern is to re-use fields between logrusging statements by re-using
//// the logrusrus.Entry returned from WithFields()
//contextLogger := logrus.WithFields(logrus.Fields{
//"common": "this is a common field",
//"other":  "I also should be logrusged always",
//})
//
//contextLogger.Info("I'll be logrusged with common and other field")
//contextLogger.Info("Me too")
//
//logrus.WithFields(logrus.Fields{
//"omg":    true,
//"number": 100,
//}).Fatal("The ice breaks!")
