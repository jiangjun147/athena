package mock

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/redis"
)

func InitRedisCli() {
	s, err := miniredis.Run()
	common.AssertError(err)

	cli := redis.NewRedisClient(s.Addr(), "", "")

	redisConf := config.GetValue("redis").ToMap()
	for name := range redisConf {
		redis.SetDB(name.(string), cli)
	}
}
