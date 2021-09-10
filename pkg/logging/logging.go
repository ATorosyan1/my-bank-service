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
	ServiceHider    = "API-SERVICE"
	LogDir          = "logs"
	DirPermission   = 0777
	FileName        = "api_service"
	TimeFieldKey    = "@timestamp"
	MessageFieldKey = "message"
	FileFieldKey    = "service"
)

// writerHook is a hook that writes logs of specified LogLevels to specified Writer
type writerHook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
	Formatter logrus.Formatter
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
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

// Levels define on which log levels this hook would trigger
func (hook *writerHook) Levels() []logrus.Level {
	return hook.LogLevels
}

type Configuration struct {
	Filename               string       `yaml:"fileName" json:"file_name"`
	MaxSize                int          `yaml:"maxSize" json:"max_size"`
	MaxBackups             int          `yaml:"maxBackups" json:"max_backups"`
	MaxAge                 int          `yaml:"maxAge" json:"max_age"`
	Level                  logrus.Level `yaml:"level" json:"level"`
	TimestampFormat        string       `yaml:"timestampFormat" json:"timestamp_format"`
	DisableLevelTruncation bool         `yaml:"disableLevelTruncation" json:"disable_level_truncation"`
	DisableColors          bool         `yaml:"disableColors" json:"disable_colors"`
	FullTimestamp          bool         `yaml:"fullTimestamp" json:"full_timestamp"`
	ForceColors            bool         `yaml:"forceColors" json:"force_colors"`
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

// Init initializes the logger
func Init(conf *Configuration) {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), " [API SERVICE]"
		},
		TimestampFormat:        conf.TimestampFormat,
		DisableLevelTruncation: conf.DisableLevelTruncation,
		DisableColors:          conf.DisableColors,
		FullTimestamp:          conf.FullTimestamp,
		ForceColors:            conf.ForceColors,
	}

	err := os.MkdirAll(LogDir, DirPermission)

	if err != nil || os.IsExist(err) {
		panic("can't create log dir. no configured logging to files")
	} else {

		infoFileHook, _ := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename:   fmt.Sprintf(conf.Filename, FileName),
			MaxSize:    conf.MaxSize, // megabytes
			MaxBackups: conf.MaxBackups,
			MaxAge:     conf.MaxAge, //days
			Level:      conf.Level,
			Formatter: &logrus.JSONFormatter{
				CallerPrettyfier: func(f *runtime.Frame) (string, string) {
					filename := path.Base(f.File)
					return fmt.Sprintf("%s:%d", filename, f.Line), ServiceHider
				},
				TimestampFormat: conf.TimestampFormat,
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
