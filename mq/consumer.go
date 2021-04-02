package mq

import (
	"context"
	"encoding/base64"
	"log"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/errcode"
	"github.com/rickone/athena/grpcex"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
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
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		reqId := uuid.New().String()
		ctx = grpcex.NewCtxWithValue(ctx, c.channel, "", "", reqId, "", "", topic)
		start := time.Now()

		defer func() {
			latency := time.Now().Sub(start).Milliseconds()

			fields := map[string]interface{}{
				"body":    base64.StdEncoding.EncodeToString(m.Body),
				"latency": latency,
			}

			if err != nil {
				fields["err"] = err.Error()
				grpcex.GetLogger(ctx).WithFields(fields).Error("Consume failed")
			} else {
				grpcex.GetLogger(ctx).WithFields(fields).Error("Consume success")
			}
		}()

		defer func() {
			if ret := recover(); ret != nil {
				stack := string(debug.Stack())
				grpcex.GetLogger(ctx).WithFields(logrus.Fields{
					"stack": stack,
					"err":   ret,
				}).Error("Recover panic")
				log.Printf("panic: %v\n%s\n", ret, stack)

				if retErr, ok := ret.(error); ok {
					err = retErr
				} else {
					err = status.Errorf(errcode.ErrRpcPanic, "recover panic: %v", ret)
				}
			}
		}()

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
