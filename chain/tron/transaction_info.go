package tron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rickone/athena/jsonex"
	"github.com/shopspring/decimal"
)

type TransactionInfo struct {
	ContractResult []string `json:"contractResult"`
	Receipt        struct {
		Result    string          `json:"result"`
		EnergyFee decimal.Decimal `json:"energy_fee"`
		NetFee    decimal.Decimal `json:"net_fee"`
	} `json:"receipt"`
}

func GetTransactionInfoById(apiUrl string, id string) (*TransactionInfo, error) {
	// https://cn.developers.tron.network/reference#gettransactioninfobyblocknum-1

	data, err := json.Marshal(map[string]interface{}{
		"value": id,
	})
	if err != nil {
		return nil, err
	}

	r, err := http.Post(fmt.Sprintf("%s/wallet/gettransactioninfobyid", apiUrl), "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	ti := TransactionInfo{}
	err = jsonex.UnmarshalFromReader(r.Body, &ti)
	return &ti, err
}
