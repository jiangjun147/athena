package tron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/rickone/athena/config"
)

type TronClient struct {
	apiUrl string
}

var (
	client *TronClient
	once   = sync.Once{}
)

func Client() *TronClient {
	once.Do(func() {
		apiUrl := config.GetString("tron", "api_url")

		client = &TronClient{
			apiUrl: apiUrl,
		}
	})
	return client
}

func (cli *TronClient) httpPost(path string, in interface{}, out interface{}) error {
	url := cli.apiUrl + path

	req, err := json.Marshal(in)
	if err != nil {
		return err
	}

	r, err := http.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	log.Printf("[TRON] POST %s \treq=%s\tresp=%s\n", url, req, resp)

	return json.Unmarshal(resp, out)
}

func (cli *TronClient) httpGet(path string, in *url.Values, out interface{}) error {
	url := cli.apiUrl + path
	req := in.Encode()

	r, err := http.Get(fmt.Sprintf("%s?%s", url, req))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	log.Printf("[TRON] GET %s \treq=%s\tresp=%s\n", url, req, resp)

	return json.Unmarshal(resp, out)
}
