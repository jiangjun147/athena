package config

import (
	"log"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/rickone/athena/common"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	errInterval = 2 * time.Second
)

func Watch() {
	address := GetString("service", "consul")
	if address == "" {
		log.Fatalf("service.consul is empty!")
	}

	cfg := api.DefaultConfig()
	cfg.Address = address

	client, err := api.NewClient(cfg)
	common.AssertError(err)

	var lastIndex uint64
	for {
		data, err := getKV(client, "config", &lastIndex)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err.Error(),
			}).Error("Config watch failed")
			time.Sleep(errInterval)
		}

		val := map[interface{}]interface{}{}
		err = yaml.Unmarshal(data, &val)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err.Error(),
			}).Error("Config watch unmarshal failed")
			time.Sleep(errInterval)
		}

		for k, v := range val {
			UpdateValue(k, v)
		}
	}
}

func getKV(client *api.Client, key string, lastIndex *uint64) ([]byte, error) {
	kvPair, meta, err := client.KV().Get(key, &api.QueryOptions{WaitIndex: *lastIndex})
	if err != nil {
		return nil, err
	}
	*lastIndex = meta.LastIndex

	if kvPair == nil {
		return nil, nil
	}

	return kvPair.Value, nil
}
