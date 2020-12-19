package logger

import (
	"encoding/json"
	"net"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type UdpGelfHook struct {
	levels []logrus.Level
	host   string
	conn   net.Conn
}

func (hook *UdpGelfHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *UdpGelfHook) Fire(entry *logrus.Entry) error {
	payload := map[string]interface{}{
		"version":       "1.1",
		"host":          hook.host,
		"timestamp":     time.Now().Unix(),
		"level":         uint32(entry.Level),
		"short_message": entry.Message, // TODO full_message include stack traceback
	}

	file, line := getFileAndLine()
	if file != "" {
		payload["_file"] = file
		payload["_line"] = line
	}

	for field, value := range entry.Data {
		payload["_"+field] = value
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = hook.conn.Write(data)
	return err
}

func NewUdpGelfHook(address string, levels ...logrus.Level) (*UdpGelfHook, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("udp", address)
	if err != nil {
		return nil, err
	}

	return &UdpGelfHook{
		levels: levels,
		host:   host,
		conn:   conn,
	}, nil
}
