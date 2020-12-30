package tron

import (
	"encoding/hex"
	"math/big"
)

type TransactionInfo struct {
	Result         string   `json:"result"`
	ResMessage     string   `json:"resMessage"`
	BlockTimestamp uint64   `json:"blockTimeStamp"`
	Fee            *big.Int `json:"fee"`
}

func (cli *TronClient) GetTransactionInfoById(id string) (*TransactionInfo, error) {
	ti := TransactionInfo{}
	err := cli.httpPost("/wallet/gettransactioninfobyid", map[string]interface{}{
		"value":   id,
		"visible": true,
	}, &ti)
	return &ti, err
}

func (ti *TransactionInfo) GetResMessage() string {
	if ti.ResMessage == "" {
		return ""
	}

	data, _ := hex.DecodeString(ti.ResMessage)
	return string(data)
}
