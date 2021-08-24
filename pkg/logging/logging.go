package logging

import (
	"fmt"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"io"
	"os"
	"path"
	"runtime"
)

const (
	ServiceHider    = "BANK_API"
	LogDir          = "logs"
	DirPermission   = 0777
	FileName        = "bank"
	TimeFieldKey    = "@timestamp"
	MessageFieldKey = "message"
	FileFieldKey    = "service"
)

type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
	Formatter logrus.Formatter
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range hook.Writer {
		_, err = w.Write([]byte(line))
	}
	return err
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

var e *logrus.Entry

type Logger struct {
	*logrus.Entry
}

func GetLogger() Logger {
	return Logger{e}
}

func (l *Logger) GetLoggerWithField(k string, v interface{}) Logger {
	return Logger{l.WithField(k, v)}
}

func Init() {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), " [BANK SERVICE]"
		},
		TimestampFormat:        "2006/01/02 - 15:04:05",
		DisableLevelTruncation: true,
		DisableColors:          false,
		FullTimestamp:          true,
		ForceColors:            true,
	}

	err := os.MkdirAll(LogDir, DirPermission)

	if err != nil || os.IsExist(err) {
		panic("can't create log dir. no configured logging to files")
	} else {

		infoFileHook, _ := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   fmt.Sprintf("./logs/%s.log", FileName),
			MaxSize:    50, // megabytes
			MaxBackups: 7,
			MaxAge:     7, //days
			Level:      5,
			Formatter: &logrus.JSONFormatter{
				CallerPrettyfier: func(f *runtime.Frame) (string, string) {
					filename := path.Base(f.File)
					return fmt.Sprintf("%s:%d", filename, f.Line), ServiceHider
				},
				TimestampFormat: "2006/01/02 - 15:04:05",
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyTime: TimeFieldKey,
					logrus.FieldKeyMsg:  MessageFieldKey,
					logrus.FieldKeyFile: FileFieldKey,
				},
			},
		})
		l.AddHook(infoFileHook)
	}

	l.SetLevel(logrus.TraceLevel)
	logrus.SetOutput(colorable.NewColorableStdout())

	e = logrus.NewEntry(l)
}
