package mysql

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
)

var (
	clients = map[string]*gorm.DB{}
	mu      = sync.RWMutex{}
)

func NewClient(addr, username, password, db string) *gorm.DB {
	cli, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, addr, db))
	common.AssertError(err)

	cli.SingularTable(true)
	return cli.Debug()
}

func DB(name string) *gorm.DB {
	cli := getMySQLCli(name)
	if cli != nil {
		return cli
	}
	return initMySQLCli(name)
}

func SetDB(name string, cli *gorm.DB) {
	clients[name] = cli
}

func initMySQLCli(name string) *gorm.DB {
	mu.Lock()
	defer mu.Unlock()

	cli, ok := clients[name]
	if ok {
		return cli
	}

	conf := config.GetValue("mysql", name)

	cli = NewClient(conf.GetString("address"), conf.GetString("username"), conf.GetString("password"), conf.GetString("db"))
	SetDB(name, cli)
	return cli
}

func getMySQLCli(name string) *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()

	return clients[name]
}
