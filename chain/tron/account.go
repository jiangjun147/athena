package tron

import (
	"github.com/shopspring/decimal"
)

type Account struct {
	Address string          `json:"address"`
	Balance decimal.Decimal `json:"balance"`
}

func (cli *TronClient) GetAccount(address string) (*Account, error) {
	acc := Account{}
	err := cli.httpPost("/wallet/getaccount", map[string]interface{}{
		"address": address,
		"visible": "true",
	}, &acc)
	return &acc, err
}
