package grpcex

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

const (
	dialTimeout = 3 * time.Second
)

var (
	clients = map[string]*grpc.ClientConn{}
	mu      = sync.RWMutex{}
)

func Dial(address string, mws ...grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	mws = append([]grpc.UnaryClientInterceptor{CtxUnaryClientMW()}, mws...)

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	return grpc.DialContext(ctx, address,
		grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithChainUnaryInterceptor(mws...),
	)
}

func DialByName(target string, mws ...grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	if os.Getenv("ENV") != "test" {
		mws = append([]grpc.UnaryClientInterceptor{RerouteUnaryClientMW(target)}, mws...)
	}

	consulAddress := config.GetString("service", "consul")
	return Dial(fmt.Sprintf("consul://%s/%s", consulAddress, target), mws...)
}

func initGrpcConn(name string) *grpc.ClientConn {
	mu.Lock()
	defer mu.Unlock()

	conn, ok := clients[name]
	if ok {
		return conn
	}

	conn, err := DialByName(name)
	common.AssertError(err)

	clients[name] = conn
	return conn
}

func getGrpcConn(name string) *grpc.ClientConn {
	mu.RLock()
	defer mu.RUnlock()

	return clients[name]
}

func Client(name string) *grpc.ClientConn {
	conn := getGrpcConn(name)
	if conn != nil {
		return conn
	}
	return initGrpcConn(name)
}
