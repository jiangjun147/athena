package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/grpcex"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	address string
	channel string
	clients []*nsq.Consumer
}

func NewConsumer(channel string) *Consumer {
	address := config.GetString("service", "nsqlookupd")

	return &Consumer{
		address: address,
		channel: channel,
	}
}

func (c *Consumer) OnWithTimeout(topic string, f func(ctx context.Context, m *nsq.Message) error, timeout time.Duration) {
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

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		reqId := uuid.New().String()
		ctx = grpcex.NewCtxWithValue(ctx, c.channel, "", "", reqId, "", "", topic)

		err = f(ctx, m)
		return
	}))

	err = consumer.ConnectToNSQLookupd(c.address)
	common.AssertError(err)

	c.clients = append(c.clients, consumer)
}

func (c *Consumer) On(topic string, f func(ctx context.Context, m *nsq.Message) error) {
	c.OnWithTimeout(topic, f, 8*time.Second)
}

func (c *Consumer) Stop() {
	for _, cli := range c.clients {
		cli.Stop()
	}
}
