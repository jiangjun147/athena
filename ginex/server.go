package ginex

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/consul"
	"github.com/rickone/athena/errcode"
	"github.com/rickone/athena/metrics"
	"google.golang.org/grpc/status"
)

type GinService struct {
	*gin.Engine
	ip4      string
	port     int
	register *consul.Register
}

func NewGinService(name string) *GinService {
	service := &GinService{}
	service.Engine = gin.New()

	address := config.GetString("service", name)

	ip4, port, err := common.AddressToIp4Port(address)
	if err != nil {
		panic(err)
	}

	service.ip4 = ip4
	if service.ip4 == "" {
		service.ip4 = common.GetLocalAddr()
	}
	service.port = port

	os.Setenv("Service", name)

	go metrics.ReportInfluxDBV2(name)

	return service
}

func (s *GinService) Register(name string) {
	s.GET("health", func(c *gin.Context) {
		c.Status(208)
	})

	cs := &consul.Register{
		Service: name,
		Address: s.ip4,
		Port:    s.port,
		Check: &api.AgentServiceCheck{
			Interval:                       "10s",
			DeregisterCriticalServiceAfter: "1m",
			HTTP:                           fmt.Sprintf("http://%s:%d/health", s.ip4, s.port),
		},
	}

	consulAddress := config.GetString("service", "consul")
	common.AssertError(cs.Register(consulAddress))
	s.register = cs
}

func (s *GinService) Serve() {
	s.Run(fmt.Sprintf(":%d", s.port))
	//r.RunTLS(fmt.Sprintf(":%d", port), "./server.crt", "./server.key")
}

func (s *GinService) Close() {
	if s.register != nil {
		consulAddress := config.GetString("service", "consul")
		s.register.Deregister(consulAddress)
		s.register = nil
	}
}

func GetBearerAccessToken(c *gin.Context) string {
	const prefix = "Bearer "
	auth := c.GetHeader("Authorization")
	token := ""

	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	}
	return token
}

func ParamInt(c *gin.Context, key string) (int64, error) {
	str := c.Param(key)

	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, status.Errorf(errcode.ErrGinParam, "param :%s need integer", key)
	}

	return val, nil
}

func ParamGlcId(c *gin.Context, key string) (int64, error) {
	str := c.Param(key)

	val, err := common.GlcDecode(str)
	if err != nil {
		return 0, status.Errorf(errcode.ErrGlcDecode, "param %s invalid", key)
	}

	return val, nil
}
