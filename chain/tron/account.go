package tron

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rickone/athena/jsonex"
	"github.com/shopspring/decimal"
)

type Account struct {
	Address string          `json:"address"`
	Balance decimal.Decimal `json:"balance"`
}

func GetAccount(apiUrl string, address string) (*Account, error) {
	// https://cn.developers.tron.network/reference#%E8%8E%B7%E5%8F%96%E5%B8%90%E6%88%B7

	data, err := json.Marshal(map[string]interface{}{
		"address": address,
		"visible": "true",
	})
	if err != nil {
		return nil, err
	}

	r, err := http.Post(fmt.Sprintf("%s/wallet/getaccount", apiUrl), "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	acc := Account{}
	err = jsonex.UnmarshalFromReader(r.Body, &acc)
	return &acc, err
}
