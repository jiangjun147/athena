package tron

import (
	"fmt"
	"net/url"
	"strconv"
)

type TransactionRecord struct {
	TxId           string `json:"transaction_id"`
	From           string `json:"from"`
	To             string `json:"to"`
	Type           string `json:"type"`
	Value          string `json:"value"`
	BlockTimestamp uint64 `json:"block_timestamp"`
}

func (cli *TronClient) GetTrc20TransactionRecord(contract string, from, to string, maxBlockTimestamp int64, trxType string) ([]*TransactionRecord, error) {
	vals := url.Values{}
	vals.Set("only_confirmed", "true")
	vals.Set("limit", "50")
	vals.Set("contract_address", contract)
	vals.Set("order_by", "block_timestamp,asc")

	var path string
	if from != "" {
		vals.Set("only_from", "true")
		path = fmt.Sprintf("/v1/accounts/%s/transactions/trc20", from)
	} else {
		vals.Set("only_to", "true")
		path = fmt.Sprintf("/v1/accounts/%s/transactions/trc20", to)
	}

	if maxBlockTimestamp > 0 {
		vals.Set("min_timestamp", strconv.FormatInt(maxBlockTimestamp+1, 10))
	}

	var result []*TransactionRecord
	for {
		resp := struct {
			Data []*TransactionRecord `json:"data"`
			Meta struct {
				Fingerprint string `json:"fingerprint"`
			} `json:"meta"`
		}{}
		err := cli.httpGet(path, &vals, &resp)
		if err != nil {
			return nil, err
		}

		for _, rec := range resp.Data {
			if from != "" && from != rec.From {
				continue
			}

			if to != "" && to != rec.To {
				continue
			}

			if trxType != "" && trxType != rec.Type {
				continue
			}

			result = append(result, rec)
		}

		fingerprint := resp.Meta.Fingerprint
		if fingerprint == "" {
			break
		}
		vals.Set("fingerprint", fingerprint)
	}
	return result, nil
}
