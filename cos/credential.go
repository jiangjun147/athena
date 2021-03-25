package cos

import (
	"net/http"
	"time"

	"github.com/rickone/athena/config"
	"github.com/rickone/athena/errcode"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"google.golang.org/grpc/status"
)

const (
	expireIn = 10 * time.Minute
)

var (
	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}
)

func CreateCredential(name string) (*sts.CredentialResult, error) {
	conf := config.GetValue("cos", name)
	if conf == nil {
		return nil, status.Error(errcode.ErrConfigNotFound, "config not found")
	}

	cli := sts.NewClient(conf.GetString("secret_id"), conf.GetString("secret_key"), httpClient)
	opt := &sts.CredentialOptions{
		DurationSeconds: int64(expireIn.Seconds()),
		Region:          "ap-guangzhou",
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					Action: []string{
						"name/cos:PostObject",
						"name/cos:PutObject",
					},
					Effect: "allow",
					Resource: []string{
						"*",
					},
				},
			},
		},
	}
	return cli.GetCredential(opt)
}
