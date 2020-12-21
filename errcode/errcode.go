package errcode

import (
	"google.golang.org/grpc/status"
)

// [http status 3dig] + [biz code 3dig]

const (
	ErrGinBind   = 400900 + iota // 绑定参数错误
	ErrGinParam                  // Url参数错误
	ErrGlcDecode                 // Glc解码
)

const (
	ErrRequestLimit  = 403900 + iota // 请求太频繁
	ErrMutexLock                     // 抢锁失败
	ErrMutexUnlock                   // 释放锁失败
	ErrBlocked                       // 系统己阻断
	ErrKeyDuplicated                 // 键冲突
)

const (
	ErrRecordNotFound = 404900 + iota // MySQL记录不存在
	ErrValueNotFound                  // Redis值不存在
)

const (
	ErrRpcTimeout = 500900 + iota // Rpc调用超时
	ErrRpcFailed                  // Rpc失败
	ErrRpcPanic                   // Rpc panic
)

func From(err error) (code int, failed bool) {
	if err != nil {
		if st, ok := status.FromError(err); ok {
			code = int(st.Code())
		} else {
			code = 500000
		}
	}

	if code != 0 && (code < 400000 || code >= 500000) {
		failed = true
	}
	return
}
