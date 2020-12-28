package eth

import (
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
)

type EthClient struct {
	*ethclient.Client
}

var (
	client *EthClient
	once   = sync.Once{}
)

func Client() *EthClient {
	once.Do(func() {
		apiUrl := config.GetString("eth", "api_url")

		cli, err := rpc.Dial(apiUrl)
		common.AssertError(err)

		client = &EthClient{
			Client: ethclient.NewClient(cli),
		}
	})
	return client
}
