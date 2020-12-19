package logger

import (
	"io"
	"path"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

const (
	rotationTime  = 1 * time.Hour
	rotationCount = 3 * 24
)

type FileRotateHook struct {
	levels []logrus.Level
	writer io.Writer
}

func (hook *FileRotateHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *FileRotateHook) Fire(entry *logrus.Entry) error {
	str, err := entry.String()
	if err != nil {
		return err
	}

	_, err = hook.writer.Write([]byte(str))
	return err
}

func NewFileRotateHook(fileName string, levels ...logrus.Level) (*FileRotateHook, error) {
	logPath, err := filepath.Abs("./log")
	if err != nil {
		return nil, err
	}

	filePath := path.Join(logPath, fileName)
	writer, err := rotatelogs.New(
		filePath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(filePath),
		rotatelogs.WithRotationTime(rotationTime),
		rotatelogs.WithRotationCount(rotationCount),
	)
	if err != nil {
		return nil, err
	}

	return &FileRotateHook{
		levels: levels,
		writer: writer,
	}, nil
}
