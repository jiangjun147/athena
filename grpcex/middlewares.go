package grpcex

import (
	"context"
	"log"
	"os"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/rickone/athena/errcode"
	"github.com/rickone/athena/logger"
	"github.com/rickone/athena/metrics"
	"github.com/rickone/athena/redis"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type GrpcCtxKey string

const (
	rpcCtxKey GrpcCtxKey = "rpc_ctx"
)

var (
	regFullMethod = regexp.MustCompile("^/([^/]+)/(.+)$")
)

func CtxUnaryClientMW() grpc.UnaryClientInterceptor {
	service := os.Getenv("Service")

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		kvs := []string{"caller", service}
		if c, ok := ctx.(*gin.Context); ok {
			reqId := c.GetString("Request-Id")
			clientIp := c.ClientIP()
			kvs = append(kvs, "request_id", reqId, "client_ip", clientIp)

			userId := ""
			authInfo, ok := c.Get("AuthInfo")
			if ok {
				val := reflect.ValueOf(authInfo)
				if val.Kind() == reflect.Ptr {
					val = val.Elem()
				}

				userId = strconv.FormatInt(val.FieldByName("UserId").Int(), 10)
				kvs = append(kvs, "user_id", userId)
			}

			ctx = NewCtxWithValue(ctx, reqId, "", service, c.FullPath(), clientIp, userId)
		} else {
			reqId := GetCtxValue(ctx, "request_id")
			if reqId != nil {
				kvs = append(kvs, "request_id", reqId.(string))
			}

			clientIp := GetCtxValue(ctx, "client_ip")
			if clientIp != nil {
				kvs = append(kvs, "client_ip", clientIp.(string))
			}

			userId := GetCtxValue(ctx, "user_id")
			if userId != nil {
				kvs = append(kvs, "user_id", userId.(string))
			}
		}

		ctx = metadata.AppendToOutgoingContext(ctx, kvs...)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func CtxUnaryServerMW(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}

	reqId := ""
	vals := md.Get("request_id")
	if len(vals) > 0 {
		reqId = vals[0]
	}

	caller := ""
	vals = md.Get("caller")
	if len(vals) > 0 {
		caller = vals[0]
	}

	clientIp := ""
	vals = md.Get("client_ip")
	if len(vals) > 0 {
		clientIp = vals[0]
	}

	userId := ""
	vals = md.Get("user_id")
	if len(vals) > 0 {
		userId = vals[0]
	}

	subs := regFullMethod.FindStringSubmatch(info.FullMethod)
	if len(subs) != 3 {
		return handler(ctx, req)
	}

	ctx = NewCtxWithValue(ctx, reqId, caller, subs[1], subs[2], clientIp, userId)
	return handler(ctx, req)
}

func NewCtxWithValue(ctx context.Context, requestId string, caller string, service string, method string, clientIp string, userId string) context.Context {
	fields := map[string]interface{}{
		"request_id": requestId,
		"service":    service,
		"method":     method,
		"client_ip":  clientIp,
	}
	if caller != "" {
		fields["caller"] = caller
	}
	if userId != "" {
		fields["user_id"] = userId
	}

	entry := logger.NewEntry(ctx, fields)

	ctxValue := map[string]interface{}{
		"request_id": requestId,
		"service":    service,
		"method":     method,
		"client_ip":  clientIp,
		"logger":     entry,
	}
	if caller != "" {
		ctxValue["caller"] = caller
	}
	if userId != "" {
		ctxValue["user_id"] = userId
	}

	return context.WithValue(ctx, rpcCtxKey, ctxValue)
}

func GetCtxValue(ctx context.Context, field string) interface{} {
	obj := ctx.Value(rpcCtxKey)
	if obj == nil {
		return nil
	}

	fields, ok := obj.(map[string]interface{})
	if !ok {
		return nil
	}

	return fields[field]
}

func GetLogger(ctx context.Context) *logrus.Entry {
	entry := GetCtxValue(ctx, "logger")
	if entry == nil {
		return nil
	}

	if obj, ok := entry.(*logrus.Entry); ok {
		return obj
	}

	return logrus.StandardLogger().WithContext(ctx)
}

func getRerouteClientConn(ctx context.Context, target string) (*grpc.ClientConn, error) {
	reqId := GetCtxValue(ctx, "request_id")
	if reqId == nil {
		return nil, nil
	}

	ss := strings.Split(reqId.(string), "#")
	if len(ss) != 2 {
		return nil, nil
	}

	addr, err := redigo.String(redis.DB("reroute").Do("HGET", ss[1], target))
	if err != nil {
		if err == redigo.ErrNil {
			return nil, nil
		}
		return nil, err
	}

	return Dial(addr)
}

func RerouteUnaryClientMW(target string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		conn, err := getRerouteClientConn(ctx, target)
		if err != nil {
			return err
		}
		if conn == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		return invoker(ctx, method, req, reply, conn, opts...)
	}
}

func AccessLogMW(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	service := GetCtxValue(ctx, "service")
	if service == nil || service.(string) == "grpc.health.v1.Health" {
		return handler(ctx, req)
	}

	start := time.Now()
	resp, err := handler(ctx, req)
	latency := time.Now().Sub(start).Milliseconds()

	fields := map[string]interface{}{
		"req":     req,
		"latency": latency,
	}

	code, failed := errcode.From(err)
	fields["code"] = code

	if err != nil {
		fields["err"] = err.Error()
	}

	logger := GetLogger(ctx).WithFields(fields)
	if code == 0 {
		logger.Info("Access success")
	} else if failed {
		logger.Error("Access failed")
	} else {
		logger.Warn("Access denied")
	}
	return resp, err
}

func RecoveryMW(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if ret := recover(); ret != nil {
			stack := string(debug.Stack())
			GetLogger(ctx).WithFields(logrus.Fields{
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
	return handler(ctx, req)
}

func MetricsUnaryMW(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	service := GetCtxValue(ctx, "service")
	if service == nil || service.(string) == "grpc.health.v1.Health" {
		return handler(ctx, req)
	}

	ts := time.Now()
	latency := metrics.NewHistogram("latency", "method", info.FullMethod)
	resp, err := handler(ctx, req)
	latency.Update(time.Since(ts).Nanoseconds())

	code, failed := errcode.From(err)
	status := "success"
	if failed {
		status = "failed"
	}

	call := metrics.NewCounter("call", "method", info.FullMethod, "status", status, "code", strconv.Itoa(code))
	call.Inc(1)
	return resp, err
}

func ErrorMapUnaryMW(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	return resp, errcode.ErrorMap(err)
}

func TimeoutUnaryMW(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := handler(newCtx, req)
		if newCtx.Err() == context.DeadlineExceeded {
			return nil, status.Error(errcode.ErrRpcTimeout, "rpc timeout")
		}
		return resp, err
	}
}
