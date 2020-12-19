package consul

import (
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"
)

type Register struct {
	Service string
	Address string
	Port    int
	Check   *api.AgentServiceCheck
}

func (r *Register) serviceID() string {
	return fmt.Sprintf("%s-%s-%d", r.Service, r.Address, r.Port)
}

func (r *Register) Register(address string) error {
	cfg := api.DefaultConfig()
	cfg.Address = address

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	log.Printf("Consul register: %+v\n", *r)
	return client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      r.serviceID(), // 服务节点的名称
		Name:    r.Service,     // 服务名称
		Tags:    []string{},    // tag，可以为空
		Port:    r.Port,        // 服务端口
		Address: r.Address,     // 服务 IP
		Check:   r.Check,
	})
}

func (r *Register) Deregister(address string) error {
	cfg := api.DefaultConfig()
	cfg.Address = address

	client, err := api.NewClient(cfg)
	if err != nil {
		return err
	}

	log.Printf("Consul deregister: %+v\n", *r)
	return client.Agent().ServiceDeregister(r.serviceID())
}
