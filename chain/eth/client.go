package eth

import (
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rickone/athena/common"
)

type EthClient struct {
	*ethclient.Client
	raw *rpc.Client
}

var (
	clients map[string]*EthClient
	mu      = sync.RWMutex{}
)

func initEthClient(apiUrl string) *EthClient {
	mu.Lock()
	defer mu.Unlock()

	cli, ok := clients[apiUrl]
	if ok {
		return cli
	}

	conn, err := rpc.Dial(apiUrl)
	common.AssertError(err)

	cli = &EthClient{
		Client: ethclient.NewClient(conn),
		raw:    conn,
	}

	clients[apiUrl] = cli
	return cli
}

func getEthClient(apiUrl string) *EthClient {
	mu.RLock()
	defer mu.RUnlock()

	return clients[apiUrl]
}

func Client(apiUrl string) *EthClient {
	cli := getEthClient(apiUrl)
	if cli != nil {
		return cli
	}
	return initEthClient(apiUrl)
}
