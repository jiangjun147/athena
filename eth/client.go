package eth

import (
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
)

var (
	client *rpc.Client
	once   = sync.Once{}
)

func RawClient() *rpc.Client {
	once.Do(func() {
		apiUrl := config.GetString("eth", "api_url")

		var err error
		client, err = rpc.Dial(apiUrl)
		common.AssertError(err)
	})
	return client
}

func Client() *ethclient.Client {
	return ethclient.NewClient(RawClient())
}
