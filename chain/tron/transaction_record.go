package tron

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/rickone/athena/jsonex"
	"github.com/shopspring/decimal"
)

type TransactionRecord struct {
	TxId           string          `json:"transaction_id"`
	From           string          `json:"from"`
	To             string          `json:"to"`
	Type           string          `json:"type"`
	Value          decimal.Decimal `json:"value"`
	BlockTimestamp int64           `json:"block_timestamp"`
}

func GetTrc20TransactionRecord(apiUrl string, contract string, from, to string, lastTimestamp int64, trxType string) ([]*TransactionRecord, error) {
	vals := url.Values{}
	vals.Add("only_confirmed", "true")
	vals.Add("contract_address", contract)
	vals.Add("order_by", "block_timestamp,asc")

	var location string
	if from != "" {
		vals.Add("only_from", "true")
		location = fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20", apiUrl, from)
	} else {
		vals.Add("only_to", "true")
		location = fmt.Sprintf("%s/v1/accounts/%s/transactions/trc20", apiUrl, to)
	}

	if lastTimestamp > 0 {
		vals.Add("min_timestamp", strconv.FormatInt(lastTimestamp+1, 10))
	}

	r, err := http.Get(fmt.Sprintf("%s?%s", location, vals.Encode()))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	resp := struct {
		Data []*TransactionRecord `json:"data"`
	}{}
	if err := jsonex.UnmarshalFromReader(r.Body, &resp); err != nil {
		return nil, err
	}

	var result []*TransactionRecord
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
	return result, nil
}
