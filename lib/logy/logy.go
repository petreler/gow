package logy

import (
	rotate "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

// LoggerConfig  日志配置
type LoggerConfig struct {
	Path       string //文件目录
	FileName   string //文件名
	ToFile     bool   //是否保存到文件
	MaxDay     int    //最大保存天数
	SplitDay   int    //切割天数
	TimeFormat string //时间格式
}

var (
	logger            = logrus.New()
	defaultWriter     = os.Stdout
	defaultTimeFormat = "2006-01-02 15:04:05"
	defaultMaxDay     = 7
	defaultSplitDay   = 1
)

//InitLogger 要写文件时，记得init此方法
func InitLogger(conf *LoggerConfig) {
	if conf.MaxDay == 0 {
		conf.MaxDay = defaultMaxDay
	}
	if conf.SplitDay == 0 {
		conf.SplitDay = defaultSplitDay
	}
	if conf.TimeFormat == "" {
		conf.TimeFormat = defaultTimeFormat
	}

	if conf.ToFile && (conf.Path == "" || conf.FileName == "") {
		panic("请设置要保存的日志文件参数")
	}

	if conf.ToFile {
		fileName := path.Join(conf.Path, conf.FileName)
		src, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			panic(err)
		}

		logger.Out = src

		//文件生成格式
		logWriter, err := rotate.New(
			fileName+"_%Y%m%d.log",
			rotate.WithLinkName(fileName),
			rotate.WithMaxAge(time.Duration(conf.MaxDay)*24*time.Hour),
			rotate.WithRotationTime(time.Duration(conf.SplitDay)*24*time.Hour),
		)

		writeMap := lfshook.WriterMap{
			logrus.InfoLevel:  logWriter,
			logrus.FatalLevel: logWriter,
			logrus.DebugLevel: logWriter,
			logrus.WarnLevel:  logWriter,
			logrus.ErrorLevel: logWriter,
			logrus.PanicLevel: logWriter,
		}

		lfHook := lfshook.NewHook(writeMap, &logrus.TextFormatter{
			TimestampFormat: conf.TimeFormat,
		})
		logger.AddHook(lfHook)
	}
}

// default init
func init() {
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: defaultTimeFormat,
		FullTimestamp:   true,
	})
	logger.Out = defaultWriter
	logger.SetLevel(logrus.DebugLevel)
	logger.SetReportCaller(false)
}

// Info
func Info(v ...interface{}) {
	logger.Info(v...)
}

// Debug
func Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Error
func Error(v ...interface{}) {
	logger.Error(v...)
}

// Trace
func Trace(v ...interface{}) {
	logger.Trace(v...)
}

func Panic(v ...interface{}) {
	logger.Panic(v...)
}
