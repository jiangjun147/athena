package mq

import (
	"encoding/json"
	"sync"
	"time"

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
		conf := config.GetMapValue("Service", "nsqd").(*config.ServiceConf)

		var err error
		producerCli, err = nsq.NewProducer(conf.Address, nsq.NewConfig())
		common.AssertError(err)
	})
	return producerCli
}

func Publish(topic string, value interface{}) {
	data, err := json.Marshal(value)
	common.AssertError(err)

	err = getProducer().Publish(topic, data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": topic,
			"err":   err.Error(),
		}).Error("Publish failed")
	}
}

func DeferredPublish(topic string, delay time.Duration, value interface{}) {
	data, err := json.Marshal(value)
	common.AssertError(err)

	err = getProducer().DeferredPublish(topic, delay, data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": topic,
			"delay": delay.Seconds(),
			"err":   err.Error(),
		}).Error("DeferredPublish failed")
	}
}
