package logger

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/sirupsen/logrus"
)

func Init(name string) {
	traceLogHook, err := NewFileRotateHook(fmt.Sprintf("%s/trace.log", name), logrus.WarnLevel, logrus.InfoLevel, logrus.DebugLevel, logrus.TraceLevel)
	common.AssertError(err)
	logrus.AddHook(traceLogHook)

	errorLogHook, err := NewFileRotateHook(fmt.Sprintf("%s/error.log", name), logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel)
	common.AssertError(err)
	logrus.AddHook(errorLogHook)

	graylogAddress := config.GetString("service", "graylog")
	if graylogAddress != "" {
		gelfHook, err := NewUdpGelfHook(graylogAddress, logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel, logrus.InfoLevel)
		common.AssertError(err)
		logrus.AddHook(gelfHook)
	}
	logrus.SetFormatter(&logrus.TextFormatter{
		//DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "01/02 15:04:05",
	})
}

func NewEntry(ctx context.Context, fields map[string]interface{}) *logrus.Entry {
	return logrus.WithContext(ctx).WithFields(fields)
}

func getFileAndLine() (string, int) {
	for i := 3; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		if !strings.Contains(file, "logrus") {
			return file, line
		}
	}
	return "", 0
}
