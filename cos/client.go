package cos

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/rickone/athena/errcode"
	"github.com/tencentyun/cos-go-sdk-v5"
	"google.golang.org/grpc/status"
)

var (
	clients = map[string]*cos.Client{}
	mu      = sync.RWMutex{}
)

func initCosClient(name string) *cos.Client {
	mu.Lock()
	defer mu.Unlock()

	cli, ok := clients[name]
	if ok {
		return cli
	}

	conf := config.GetValue("cos", name)
	u, err := url.Parse(conf.GetString("base_url"))
	common.AssertError(err)

	b := &cos.BaseURL{
		BucketURL: u,
	}
	cli = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  conf.GetString("secret_id"),
			SecretKey: conf.GetString("secret_key"),
		},
	})

	clients[name] = cli
	return cli
}

func getCosClient(name string) *cos.Client {
	mu.RLock()
	defer mu.RUnlock()

	return clients[name]
}

func Bucket(name string) *cos.Client {
	conn := getCosClient(name)
	if conn != nil {
		return conn
	}
	return initCosClient(name)
}

func BucketWithCredential(name string, tmpSecretId string, tmpSecretKey string, sessionToken string) (*cos.Client, error) {
	conf := config.GetValue("cos", name)
	if conf == nil {
		return nil, status.Error(errcode.ErrConfigNotFound, "config not found")
	}

	u, err := url.Parse(conf.GetString("base_url"))
	if err != nil {
		return nil, err
	}

	b := &cos.BaseURL{
		BucketURL: u,
	}
	return cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     tmpSecretId,
			SecretKey:    tmpSecretKey,
			SessionToken: sessionToken,
		},
	}), nil
}
