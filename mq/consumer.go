package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	address string
	channel string
	clients []*nsq.Consumer
}

func NewConsumer(channel string) *Consumer {
	conf := config.GetMapValue("Service", "nsqlookupd").(*config.ServiceConf)
	return &Consumer{
		address: conf.Address,
		channel: channel,
	}
}

func (c *Consumer) On(topic string, f func(ctx context.Context, m *nsq.Message) error) {
	consumer, err := nsq.NewConsumer(topic, c.channel, nsq.NewConfig())
	common.AssertError(err)

	consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) (err error) {
		defer func() {
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"topic":  topic,
					"msg_id": fmt.Sprintf("0x%x", m.ID),
					"err":    err.Error(),
				}).Error("Consume failed")
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = f(ctx, m)
		return
	}))

	err = consumer.ConnectToNSQLookupd(c.address)
	common.AssertError(err)

	c.clients = append(c.clients, consumer)
}

func (c *Consumer) Stop() {
	for _, cli := range c.clients {
		cli.Stop()
	}
}
