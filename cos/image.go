package cos

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/rickone/athena/common"
	"github.com/rickone/athena/config"
	"github.com/tencentyun/cos-go-sdk-v5"
)

var (
	imageCli  *cos.Client
	imageOnce = sync.Once{}
)

func ImageClient() *cos.Client {
	imageOnce.Do(func() {
		conf := config.GetValue("image")

		u, err := url.Parse(conf.GetString("base_url"))
		common.AssertError(err)

		b := &cos.BaseURL{
			BucketURL: u,
		}
		imageCli = cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  conf.GetString("secret_id"),
				SecretKey: conf.GetString("secret_key"),
			},
		})
	})
	return imageCli
}
