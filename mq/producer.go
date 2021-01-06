package mq

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nsqio/go-nsq"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/sirupsen/logrus"
)

var (
	producerCli  *nsq.Producer
	producerOnce = sync.Once{}
)

func getProducer() *nsq.Producer {
	producerOnce.Do(func() {
		address := config.GetString("service", "nsqd")

		var err error
		producerCli, err = nsq.NewProducer(address, nsq.NewConfig())
		common.AssertError(err)
	})
	return producerCli
}

func Publish(topic string, body []byte) {
	err := getProducer().Publish(topic, body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": topic,
			"err":   err.Error(),
		}).Error("Publish failed")
	}
}

func DeferredPublish(topic string, delay time.Duration, body []byte) {
	err := getProducer().DeferredPublish(topic, delay, body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": topic,
			"delay": delay.Seconds(),
			"err":   err.Error(),
		}).Error("DeferredPublish failed")
	}
}

func PublishJSON(topic string, value interface{}) {
	data, err := json.Marshal(value)
	common.AssertError(err)

	Publish(topic, data)
}

func DeferredPublishJSON(topic string, delay time.Duration, value interface{}) {
	data, err := json.Marshal(value)
	common.AssertError(err)

	DeferredPublish(topic, delay, data)
}

func PublishProto(topic string, value proto.Message) {
	data, err := proto.Marshal(value)
	common.AssertError(err)

	Publish(topic, data)
}

func DeferredPublishProto(topic string, delay time.Duration, value proto.Message) {
	data, err := proto.Marshal(value)
	common.AssertError(err)

	DeferredPublish(topic, delay, data)
}
