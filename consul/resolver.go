package consul

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&Builder{})
}

type Builder struct {
}

func (cb *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	log.Printf("Consul build target: %v\n", target)

	config := api.DefaultConfig()
	config.Address = target.Authority

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	cr := &Resolver{
		client:    client,
		name:      target.Endpoint,
		cc:        cc,
		lastIndex: 0,
	}

	go func() {
		for {
			err := cr.resolve()
			if err != nil {
				logrus.WithField("err", err.Error()).Error("Consul resolve failed")
				time.Sleep(2 * time.Second)
			}
		}
	}()
	return cr, nil
}

func (cb *Builder) Scheme() string {
	return "consul"
}

type Resolver struct {
	client    *api.Client
	name      string
	cc        resolver.ClientConn
	lastIndex uint64
}

func (cr *Resolver) resolve() error {
	services, meta, err := cr.client.Health().Service(cr.name, "", true, &api.QueryOptions{WaitIndex: cr.lastIndex})
	if err != nil {
		return err
	}
	cr.lastIndex = meta.LastIndex

	var newAddrs []resolver.Address
	for _, service := range services {
		addr := fmt.Sprintf("%v:%v", service.Service.Address, service.Service.Port)
		newAddrs = append(newAddrs, resolver.Address{Addr: addr})
	}

	if len(newAddrs) == 0 {
		logrus.WithField("service", cr.name).Warn("Consul resolve NONE node!")
	}

	cr.cc.UpdateState(resolver.State{Addresses: newAddrs})
	return nil
}

func (cr *Resolver) ResolveNow(opt resolver.ResolveNowOptions) {
}

func (cr *Resolver) Close() {
}
