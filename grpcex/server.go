package grpcex

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // http pprof
	"os"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/consul"
	"github.com/rickone/athena/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	rpcTimeout = 10 * time.Second
)

type GrpcService struct {
	*grpc.Server
	name     string
	register *consul.Register
	listener net.Listener
	address  string
}

func NewGrpcService(serviceName ...string) *GrpcService {
	serv := &GrpcService{
		Server: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				CtxUnaryServerMW,
				AccessLogMW,
				RecoveryMW,
				MetricsUnaryMW,
				ErrorMapUnaryMW,
				TimeoutUnaryMW(rpcTimeout),
			),
		),
	}
	if len(serviceName) > 0 {
		serv.name = serviceName[0]
	}
	return serv
}

func (s *GrpcService) Serve() {
	serviceName := s.name
	if serviceName == "" {
		for name := range s.GetServiceInfo() {
			serviceName = name
			break
		}
	}

	address := config.GetString("service", serviceName)

	ip4, port, err := common.AddressToIp4Port(address)
	if err != nil {
		panic(err)
	}

	if ip4 == "" {
		ip4 = common.GetLocalAddr()
	}
	s.address = fmt.Sprintf("%s:%d", ip4, port)

	listener, err := net.Listen("tcp", address)
	common.AssertError(err)
	s.listener = listener

	if os.Getenv("ENV") != "test" {
		grpc_health_v1.RegisterHealthServer(s.Server, new(healthService))

		register := &consul.Register{
			Service: serviceName,
			Address: ip4,
			Port:    port,
			Check: &api.AgentServiceCheck{
				Interval:                       "3s",
				DeregisterCriticalServiceAfter: "1m",
				GRPC:                           fmt.Sprintf("%s:%d/%s", ip4, port, serviceName),
			},
		}

		consulAddress := config.GetString("service", "consul")
		common.AssertError(register.Register(consulAddress))
		s.register = register
	}
	os.Setenv("Service", serviceName)

	go func() {
		addr := fmt.Sprintf("%s:%d", ip4, port+10000)
		log.Printf("Http pprof listening on: %s\n", addr)
		http.ListenAndServe(addr, nil)
	}()

	go metrics.ReportInfluxDBV2(serviceName)

	log.Printf("gRPC start serving on: %s\n", address)
	s.Server.Serve(listener)
}

func (s *GrpcService) Address() string {
	return s.address
}

func (s *GrpcService) Close() {
	if s.register != nil {
		consulAddress := config.GetString("service", "consul")
		s.register.Deregister(consulAddress)
		s.register = nil
	}

	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
}

func RecvServerStreamForever(stream grpc.ServerStream) error {
	for {
		var msg RawBuf
		if err := stream.RecvMsg(msg); err != nil {
			return err
		}
	}
}
