package redis

import (
	"fmt"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/rickone/athena/errcode"
	"google.golang.org/grpc/status"
)

// 阻断器，系统开关，用于临时阻断调用或者某些功能
func CheckBlock(module string, key string) error {
	blocked, err := redigo.Bool(DB("blocker").Do("GET", fmt.Sprintf("%s:%s", module, key)))
	if err != nil && err != redigo.ErrNil {
		return err
	}

	if blocked {
		return status.Error(errcode.ErrBlocked, "system blocked")
	}
	return nil
}

// 系统是否关闭，默认情况都是false
func IsSystemOff(module string, key string) bool {
	return CheckBlock(module, key) != nil
}
