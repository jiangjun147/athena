package cos

import (
	"net/http"
	"time"

	"github.com/rickone/athena/config"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
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
