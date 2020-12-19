package main

import (
	"github.com/gin-gonic/gin"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/ginex"
	"github.com/rickone/athena/logger"
)

func getLimiterId(c *gin.Context) string {
	if c.FullPath() == "/v1/phone_code" {
		return c.PostForm("PhoneNumber")
	}

	var id string
	val, ok := c.Get("AuthInfo")
	if !ok {
		id = c.ClientIP()
	} else {
		id = val.(*ginex.AuthInfo).GetId()
	}
	return id
}

func main() {
	common.InitRand()
	config.Init()
	logger.Init("gateway")

	r := ginex.NewGinService("trago.Gateway")
	r.Use(
		ginex.BlockerMW(),
		ginex.LimiterMW(16, 1, getLimiterId),
		ginex.RequestIdMW(),
		ginex.AccessLogMW(),
		ginex.RecoveryMW(),
		ginex.MetricsMW(),
	)

	common.OnSigQuit(r.Close)

	r.Serve()
}
