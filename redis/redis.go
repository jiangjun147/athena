package redis

import (
	"sync"
	"time"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/rickone/athena/config"
)

var (
	clients = map[string]*RedisClient{}
	mu      = sync.RWMutex{}
)

type RedisClient struct {
	*redigo.Pool
}

func NewRedisClient(addr, password string, db int64) *RedisClient {
	return &RedisClient{
		Pool: &redigo.Pool{
			MaxIdle:     10,
			IdleTimeout: 240 * time.Second,
			TestOnBorrow: func(c redigo.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			Dial: func() (redigo.Conn, error) {
				return dial("tcp", addr, password, db)
			},
		},
	}
}

func (cli *RedisClient) Do(cmd string, args ...interface{}) (interface{}, error) {
	conn := cli.Pool.Get()
	defer conn.Close()

	return conn.Do(cmd, args...)
}

func dial(network, address, password string, db int64) (redigo.Conn, error) {
	c, err := redigo.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	if db != 0 {
		if _, err := c.Do("SELECT", db); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

func DB(name string) *RedisClient {
	cli := getRedisCli(name)
	if cli != nil {
		return cli
	}
	return initRedisCli(name)
}

func SetDB(name string, cli *RedisClient) {
	clients[name] = cli
}

func initRedisCli(name string) *RedisClient {
	mu.Lock()
	defer mu.Unlock()

	cli := clients[name]
	if cli != nil {
		return cli
	}

	conf := config.GetValue("redis", name)
	if conf == nil {
		return nil
	}

	cli = NewRedisClient(conf.GetString("address"), conf.GetString("auth"), conf.GetInt("db"))
	SetDB(name, cli)
	return cli
}

func getRedisCli(name string) *RedisClient {
	mu.RLock()
	defer mu.RUnlock()

	return clients[name]
}
