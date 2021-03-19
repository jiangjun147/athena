package tron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type TronClient struct {
	apiUrl   string
	feeLimit uint64
}

var (
	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}
)

func Client(apiUrl string, feeLimit uint64) *TronClient {
	return &TronClient{
		apiUrl:   apiUrl,
		feeLimit: feeLimit,
	}
}

func (cli *TronClient) httpPost(path string, in interface{}, out interface{}) error {
	url := cli.apiUrl + path

	req, err := json.Marshal(in)
	if err != nil {
		return err
	}

	r, err := httpClient.Post(url, "application/json", bytes.NewReader(req))
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

	r, err := httpClient.Get(fmt.Sprintf("%s?%s", url, req))
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
