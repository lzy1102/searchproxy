package logs

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"searchproxy/app/fram/config"
	"sync"
)

var log *logrus.Logger
var on sync.Once

func Install() *logrus.Logger {
	on.Do(func() {
		log = logrus.New()
		if config.Install().Mod() {
			log.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: "2006-01-02 15:04:05", // 设置json里的日期输出格式
			})
			log.SetReportCaller(true)
			log.SetOutput(&lumberjack.Logger{
				Filename:   "log/log.log",
				MaxSize:    50, // M
				LocalTime:  true,
				Compress:   true,
				MaxBackups: 5,
				MaxAge:     30, //days
			})
		} else {
			log.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: "2006-01-02 15:04:05", // 设置json里的日期输出格式
			})
			log.SetReportCaller(true)
			log.Out = os.Stdout
		}
	})
	return log
}
